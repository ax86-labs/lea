package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	graph "github.com/ax86-labs/lea/internal/graph/contracts"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func NewStore(dsn string) (*Store, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	// Enable foreign keys
	// Disabled for now because call edges often reference unknown/external symbols
	// if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
	// 	return nil, err
	// }

	if err := db.Ping(); err != nil {
		return nil, err
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Store) migrate() error {
	nodesQuery := `CREATE TABLE IF NOT EXISTS nodes (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		name TEXT NOT NULL,
		file TEXT NOT NULL,
		line INTEGER NOT NULL,
		metadata TEXT
	)`
	if _, err := s.db.Exec(nodesQuery); err != nil {
		return fmt.Errorf("failed to execute migration %q: %w", nodesQuery, err)
	}

	if err := s.ensureEdgesTable(); err != nil {
		return err
	}

	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_edges_from ON edges(from_id)`,
		`CREATE INDEX IF NOT EXISTS idx_edges_to ON edges(to_id)`,
		`CREATE INDEX IF NOT EXISTS idx_edges_type ON edges(type)`,
		`CREATE INDEX IF NOT EXISTS idx_edges_sequence ON edges(sequence)`,
	}
	for _, q := range indexes {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("failed to execute migration %q: %w", q, err)
		}
	}
	return nil
}

func (s *Store) ensureEdgesTable() error {
	exists, err := s.edgesTableExists()
	if err != nil {
		return err
	}
	if !exists {
		return s.createEdgesTable()
	}

	hasSequence, err := s.edgesHasSequence()
	if err != nil {
		return err
	}
	if hasSequence {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(`ALTER TABLE edges RENAME TO edges_old`); err != nil {
		tx.Rollback()
		return err
	}
	if _, err := tx.Exec(s.edgesTableDDL()); err != nil {
		tx.Rollback()
		return err
	}
	if _, err := tx.Exec(`INSERT INTO edges (from_id, to_id, type, sequence, metadata)
		SELECT from_id, to_id, type, 0, metadata FROM edges_old`); err != nil {
		tx.Rollback()
		return err
	}
	if _, err := tx.Exec(`DROP TABLE edges_old`); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *Store) createEdgesTable() error {
	_, err := s.db.Exec(s.edgesTableDDL())
	if err != nil {
		return fmt.Errorf("failed to execute migration %q: %w", s.edgesTableDDL(), err)
	}
	return nil
}

func (s *Store) edgesTableDDL() string {
	return `CREATE TABLE IF NOT EXISTS edges (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		from_id TEXT NOT NULL,
		to_id TEXT NOT NULL,
		type TEXT NOT NULL,
		sequence INTEGER NOT NULL DEFAULT 0,
		metadata TEXT,
		UNIQUE (from_id, to_id, type, sequence),
		FOREIGN KEY (from_id) REFERENCES nodes(id) ON DELETE CASCADE,
		FOREIGN KEY (to_id) REFERENCES nodes(id) ON DELETE CASCADE
	)`
}

func (s *Store) edgesTableExists() (bool, error) {
	row := s.db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='edges'`)
	var name string
	switch err := row.Scan(&name); err {
	case nil:
		return true, nil
	case sql.ErrNoRows:
		return false, nil
	default:
		return false, err
	}
}

func (s *Store) edgesHasSequence() (bool, error) {
	rows, err := s.db.Query(`PRAGMA table_info(edges)`)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dfltValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			return false, err
		}
		if name == "sequence" {
			return true, nil
		}
	}
	return false, rows.Err()
}

func (s *Store) SaveNode(ctx context.Context, node *graph.Node) error {
	metadata, err := marshalMetadata(node.Metadata)
	if err != nil {
		return err
	}

	query := `INSERT OR REPLACE INTO nodes (id, type, name, file, line, metadata) VALUES (?, ?, ?, ?, ?, ?)`
	_, err = s.db.ExecContext(ctx, query, node.ID, node.Type, node.Name, node.File, node.Line, metadata)
	return err
}

func (s *Store) SaveEdge(ctx context.Context, edge *graph.Edge) error {
	metadata, err := marshalMetadata(edge.Metadata)
	if err != nil {
		return err
	}

	query := `INSERT OR REPLACE INTO edges (from_id, to_id, type, sequence, metadata) VALUES (?, ?, ?, ?, ?)`
	_, err = s.db.ExecContext(ctx, query, edge.FromID, edge.ToID, edge.Type, edge.Sequence, metadata)
	return err
}

func (s *Store) SaveGraph(ctx context.Context, nodes []*graph.Node, edges []*graph.Edge) error {
	if len(nodes) == 0 && len(edges) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := s.saveNodesEdgesTx(ctx, tx, nodes, edges); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s *Store) saveNodesEdgesTx(ctx context.Context, tx *sql.Tx, nodes []*graph.Node, edges []*graph.Edge) error {
	if len(nodes) > 0 {
		stmt, err := tx.PrepareContext(ctx, `INSERT OR REPLACE INTO nodes (id, type, name, file, line, metadata) VALUES (?, ?, ?, ?, ?, ?)`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, node := range nodes {
			metadata, err := marshalMetadata(node.Metadata)
			if err != nil {
				return err
			}
			if _, err := stmt.ExecContext(ctx, node.ID, node.Type, node.Name, node.File, node.Line, metadata); err != nil {
				return err
			}
		}
	}

	if len(edges) > 0 {
		stmt, err := tx.PrepareContext(ctx, `INSERT OR REPLACE INTO edges (from_id, to_id, type, sequence, metadata) VALUES (?, ?, ?, ?, ?)`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, edge := range edges {
			metadata, err := marshalMetadata(edge.Metadata)
			if err != nil {
				return err
			}
			if _, err := stmt.ExecContext(ctx, edge.FromID, edge.ToID, edge.Type, edge.Sequence, metadata); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Store) GetNode(ctx context.Context, id string) (*graph.Node, error) {
	query := `SELECT id, type, name, file, line, metadata FROM nodes WHERE id = ?`
	row := s.db.QueryRowContext(ctx, query, id)

	var node graph.Node
	var metadataStr string
	var nodeType string
	err := row.Scan(&node.ID, &nodeType, &node.Name, &node.File, &node.Line, &metadataStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	node.Type = graph.NodeType(nodeType)

	if err := unmarshalMetadata(metadataStr, &node.Metadata); err != nil {
		return nil, err
	}

	return &node, nil
}

func (s *Store) ListNodes(ctx context.Context) ([]*graph.Node, error) {
	query := `SELECT id, type, name, file, line, metadata FROM nodes`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*graph.Node
	for rows.Next() {
		var node graph.Node
		var metadataStr string
		var nodeType string
		err := rows.Scan(&node.ID, &nodeType, &node.Name, &node.File, &node.Line, &metadataStr)
		if err != nil {
			return nil, err
		}
		node.Type = graph.NodeType(nodeType)
		if err := unmarshalMetadata(metadataStr, &node.Metadata); err != nil {
			return nil, err
		}
		nodes = append(nodes, &node)
	}
	return nodes, nil
}

func (s *Store) ListEdges(ctx context.Context) ([]*graph.Edge, error) {
	query := `SELECT from_id, to_id, type, sequence, metadata FROM edges`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []*graph.Edge
	for rows.Next() {
		var edge graph.Edge
		var edgeType string
		var metadataStr string
		if err := rows.Scan(&edge.FromID, &edge.ToID, &edgeType, &edge.Sequence, &metadataStr); err != nil {
			return nil, err
		}
		edge.Type = graph.EdgeType(edgeType)
		if err := unmarshalMetadata(metadataStr, &edge.Metadata); err != nil {
			return nil, err
		}
		edges = append(edges, &edge)
	}
	return edges, nil
}

func (s *Store) GetNeighbors(ctx context.Context, id string) ([]*graph.Node, []*graph.Edge, error) {
	// Outbound edges and their target nodes
	query := `
		SELECT e.from_id, e.to_id, e.type, e.sequence, e.metadata, n.id, n.type, n.name, n.file, n.line, n.metadata
		FROM edges e
		LEFT JOIN nodes n ON e.to_id = n.id
		WHERE e.from_id = ?
	`
	rows, err := s.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var nodes []*graph.Node
	var edges []*graph.Edge

	for rows.Next() {
		var e graph.Edge
		var eType string
		var eMetadataStr string
		var nID sql.NullString
		var nType sql.NullString
		var nName sql.NullString
		var nFile sql.NullString
		var nLine sql.NullInt64
		var nMetadataStr sql.NullString

		err := rows.Scan(
			&e.FromID, &e.ToID, &eType, &e.Sequence, &eMetadataStr,
			&nID, &nType, &nName, &nFile, &nLine, &nMetadataStr,
		)
		if err != nil {
			return nil, nil, err
		}

		e.Type = graph.EdgeType(eType)
		if err := unmarshalMetadata(eMetadataStr, &e.Metadata); err != nil {
			return nil, nil, err
		}
		edges = append(edges, &e)

		node := &graph.Node{
			ID:   e.ToID, // Default to to_id if node not found
			Name: e.ToID,
		}
		if nID.Valid {
			node.ID = nID.String
			node.Type = graph.NodeType(nType.String)
			node.Name = nName.String
			node.File = nFile.String
			node.Line = int(nLine.Int64)
			if err := unmarshalMetadata(nMetadataStr.String, &node.Metadata); err != nil {
				return nil, nil, err
			}
		} else {
			node.Type = "unknown"
		}
		nodes = append(nodes, node)
	}

	return nodes, edges, nil
}

func (s *Store) GetInboundEdges(ctx context.Context, id string) ([]*graph.Node, []*graph.Edge, error) {
	query := `
		SELECT e.from_id, e.to_id, e.type, e.sequence, e.metadata, n.id, n.type, n.name, n.file, n.line, n.metadata
		FROM edges e
		JOIN nodes n ON e.from_id = n.id
		WHERE e.to_id = ?
	`
	rows, err := s.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var nodes []*graph.Node
	var edges []*graph.Edge

	for rows.Next() {
		var e graph.Edge
		var eType string
		var eMetadataStr string
		var n graph.Node
		var nType string
		var nMetadataStr string

		err := rows.Scan(
			&e.FromID, &e.ToID, &eType, &e.Sequence, &eMetadataStr,
			&n.ID, &nType, &n.Name, &n.File, &n.Line, &nMetadataStr,
		)
		if err != nil {
			return nil, nil, err
		}

		e.Type = graph.EdgeType(eType)
		if err := unmarshalMetadata(eMetadataStr, &e.Metadata); err != nil {
			return nil, nil, err
		}
		edges = append(edges, &e)

		n.Type = graph.NodeType(nType)
		if err := unmarshalMetadata(nMetadataStr, &n.Metadata); err != nil {
			return nil, nil, err
		}
		nodes = append(nodes, &n)
	}

	return nodes, edges, nil
}

func (s *Store) DeleteNode(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM nodes WHERE id = ?`, id)
	return err
}

func (s *Store) DeleteByFile(ctx context.Context, file string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM nodes WHERE file = ?`, file)
	return err
}

func (s *Store) DeleteEdgesFrom(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM edges WHERE from_id = ?`, id)
	return err
}

func (s *Store) Close() error {
	return s.db.Close()
}

func marshalMetadata(metadata map[string]interface{}) (string, error) {
	if metadata == nil {
		return "", nil
	}
	bytes, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func unmarshalMetadata(data string, out *map[string]interface{}) error {
	if data == "" {
		return nil
	}
	return json.Unmarshal([]byte(data), out)
}
