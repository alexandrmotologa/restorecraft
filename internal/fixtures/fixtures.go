package fixtures

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// GenerateAll produces valid and corrupted test backup files into targetDir.
func GenerateAll(targetDir string) error {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("create fixtures directory: %w", err)
	}

	// 1. Valid e-commerce SQLite binary database
	validDbPath := filepath.Join(targetDir, "valid_ecommerce.sqlite")
	if err := generateValidSQLite(validDbPath); err != nil {
		return fmt.Errorf("generate valid sqlite: %w", err)
	}

	// 2. Valid e-commerce gzipped SQL dump
	validGzPath := filepath.Join(targetDir, "valid_ecommerce.sql.gz")
	if err := generateValidGzipDump(validGzPath); err != nil {
		return fmt.Errorf("generate valid gzipped dump: %w", err)
	}

	// 3. Corrupt header SQLite binary
	corruptPath := filepath.Join(targetDir, "corrupt_header.sqlite")
	if err := generateCorruptHeaderSQLite(corruptPath); err != nil {
		return fmt.Errorf("generate corrupt header sqlite: %w", err)
	}

	// 4. Relational Foreign Key violation SQLite database
	fkPath := filepath.Join(targetDir, "fk_violation.sqlite")
	if err := generateFkViolationSQLite(fkPath); err != nil {
		return fmt.Errorf("generate fk violation sqlite: %w", err)
	}

	// 5. Truncated gzip stream
	truncatedPath := filepath.Join(targetDir, "truncated.sql.gz")
	if err := generateTruncatedGzip(truncatedPath); err != nil {
		return fmt.Errorf("generate truncated gzip: %w", err)
	}

	return nil
}

func generateValidSQLite(path string) error {
	_ = os.Remove(path)
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(ON)")
	if err != nil {
		return err
	}
	defer db.Close()

	schema := `
	CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sku TEXT UNIQUE NOT NULL,
		title TEXT NOT NULL,
		price_cents INTEGER NOT NULL
	);

	CREATE TABLE orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL REFERENCES users(id),
		total_cents INTEGER NOT NULL,
		status TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE order_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
		product_id INTEGER NOT NULL REFERENCES products(id),
		quantity INTEGER NOT NULL,
		unit_price_cents INTEGER NOT NULL
	);

	INSERT INTO users (email, name) VALUES
		('alice@example.com', 'Alice Smith'),
		('bob@example.com', 'Bob Jones'),
		('carol@example.com', 'Carol White'),
		('dan@example.com', 'Dan Evans'),
		('eve@example.com', 'Eve Black');

	INSERT INTO products (sku, title, price_cents) VALUES
		('PROD-001', 'Cloud Storage Node', 12900),
		('PROD-002', 'High Availability Cluster', 49900),
		('PROD-003', 'Database Proxy Agent', 7900);

	INSERT INTO orders (user_id, total_cents, status) VALUES
		(1, 12900, 'completed'),
		(1, 49900, 'processing'),
		(2, 7900, 'completed'),
		(3, 20800, 'completed'),
		(4, 49900, 'shipped');

	INSERT INTO order_items (order_id, product_id, quantity, unit_price_cents) VALUES
		(1, 1, 1, 12900),
		(2, 2, 1, 49900),
		(3, 3, 1, 7900),
		(4, 1, 1, 12900),
		(4, 3, 1, 7900),
		(5, 2, 1, 49900);
	`
	_, err = db.Exec(schema)
	return err
}

func generateValidGzipDump(path string) error {
	sqlContent := `
CREATE TABLE users (
	id INTEGER PRIMARY KEY,
	email TEXT NOT NULL,
	name TEXT NOT NULL
);

CREATE TABLE orders (
	id INTEGER PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES users(id),
	amount INTEGER NOT NULL
);

INSERT INTO users (id, email, name) VALUES (1, 'john@example.com', 'John Doe');
INSERT INTO users (id, email, name) VALUES (2, 'jane@example.com', 'Jane Doe');

INSERT INTO orders (id, user_id, amount) VALUES (101, 1, 5000);
INSERT INTO orders (id, user_id, amount) VALUES (102, 2, 12000);
`
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write([]byte(sqlContent)); err != nil {
		return err
	}
	if err := gw.Close(); err != nil {
		return err
	}

	return os.WriteFile(path, buf.Bytes(), 0644)
}

func generateCorruptHeaderSQLite(path string) error {
	// Start with a valid DB, then corrupt bytes 16 to 48
	if err := generateValidSQLite(path); err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if len(data) > 64 {
		// Overwrite page size and reserved bytes with garbage
		for i := 16; i < 48; i++ {
			data[i] = 0xFF
		}
	}

	return os.WriteFile(path, data, 0644)
}

func generateFkViolationSQLite(path string) error {
	_ = os.Remove(path)
	// Create with foreign keys initially disabled to inject orphans
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(OFF)")
	if err != nil {
		return err
	}
	defer db.Close()

	schema := `
	CREATE TABLE customers (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL
	);

	CREATE TABLE invoices (
		id INTEGER PRIMARY KEY,
		customer_id INTEGER NOT NULL REFERENCES customers(id),
		amount INTEGER NOT NULL
	);

	INSERT INTO customers (id, name) VALUES (1, 'Real Customer');
	INSERT INTO invoices (id, customer_id, amount) VALUES (10, 1, 100);
	-- Inject orphan invoice referencing non-existent customer 999
	INSERT INTO invoices (id, customer_id, amount) VALUES (11, 999, 450);
	`
	_, err = db.Exec(schema)
	return err
}

func generateTruncatedGzip(path string) error {
	sqlContent := `CREATE TABLE data (id INT); INSERT INTO data VALUES (1);`
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, _ = gw.Write([]byte(sqlContent))
	_ = gw.Close()

	raw := buf.Bytes()
	if len(raw) > 10 {
		// Cut off the last 8 bytes (trailer containing CRC32 and ISIZE)
		raw = raw[:len(raw)-8]
	}

	return os.WriteFile(path, raw, 0644)
}
