package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	graph "github.com/andev0x/ctxd/internal/graph/contracts"
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
	queries := []string{
		`CREATE TABLE IF NOT EXISTS nodes (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			name TEXT NOT NULL,
			file TEXT NOT NULL,
			line INTEGER NOT NULL,
			metadata TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS edges (
			from_id TEXT NOT NULL,
			to_id TEXT NOT NULL,
			type TEXT NOT NULL,
			metadata TEXT,
			PRIMARY KEY (from_id, to_id, type),
			FOREIGN KEY (from_id) REFERENCES nodes(id) ON DELETE CASCADE,
			FOREIGN KEY (to_id) REFERENCES nodes(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_edges_from ON edges(from_id)`,
		`CREATE INDEX IF NOT EXISTS idx_edges_to ON edges(to_id)`,
		`CREATE INDEX IF NOT EXISTS idx_edges_type ON edges(type)`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("failed to execute migration %q: %w", q, err)
		}
	}
	return nil
}

func (s *Store) SaveNode(ctx context.Context, node *graph.Node) error {
	metadata, err := json.Marshal(node.Metadata)
	if err != nil {
		return err
	}

	query := `INSERT OR REPLACE INTO nodes (id, type, name, file, line, metadata) VALUES (?, ?, ?, ?, ?, ?)`
	_, err = s.db.ExecContext(ctx, query, node.ID, node.Type, node.Name, node.File, node.Line, string(metadata))
	return err
}

func (s *Store) SaveEdge(ctx context.Context, edge *graph.Edge) error {
	metadata, err := json.Marshal(edge.Metadata)
	if err != nil {
		return err
	}

	query := `INSERT OR REPLACE INTO edges (from_id, to_id, type, metadata) VALUES (?, ?, ?, ?)`
	_, err = s.db.ExecContext(ctx, query, edge.FromID, edge.ToID, edge.Type, string(metadata))
	return err
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

	if metadataStr != "" {
		if err := json.Unmarshal([]byte(metadataStr), &node.Metadata); err != nil {
			return nil, err
		}
	}

	return &node, nil
}

func (s *Store) GetNeighbors(ctx context.Context, id string) ([]*graph.Node, []*graph.Edge, error) {
	// Outbound edges and their target nodes
	query := `
		SELECT e.from_id, e.to_id, e.type, e.metadata, n.id, n.type, n.name, n.file, n.line, n.metadata
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
			&e.FromID, &e.ToID, &eType, &eMetadataStr,
			&nID, &nType, &nName, &nFile, &nLine, &nMetadataStr,
		)
		if err != nil {
			return nil, nil, err
		}

		e.Type = graph.EdgeType(eType)
		if eMetadataStr != "" {
			json.Unmarshal([]byte(eMetadataStr), &e.Metadata)
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
			if nMetadataStr.String != "" {
				json.Unmarshal([]byte(nMetadataStr.String), &node.Metadata)
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
		SELECT e.from_id, e.to_id, e.type, e.metadata, n.id, n.type, n.name, n.file, n.line, n.metadata
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
			&e.FromID, &e.ToID, &eType, &eMetadataStr,
			&n.ID, &nType, &n.Name, &n.File, &n.Line, &nMetadataStr,
		)
		if err != nil {
			return nil, nil, err
		}

		e.Type = graph.EdgeType(eType)
		if eMetadataStr != "" {
			json.Unmarshal([]byte(eMetadataStr), &e.Metadata)
		}
		edges = append(edges, &e)

		n.Type = graph.NodeType(nType)
		if nMetadataStr != "" {
			json.Unmarshal([]byte(nMetadataStr), &n.Metadata)
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
