{{- $alias := .Aliases.Table .Table.Name -}}

// {{$alias.DownSingular}}Factory creates {{$alias.DownPlural}}
func {{$alias.DownSingular}}Factory() *db.{{$alias.DownSingular}} {
	u := &db.{{$alias.DownSingular}}{
		ID:   uuid.Must(uuid.NewV4()).String(),
	}
	return u
}

// {{$alias.DownSingular}} for persistence
type {{$alias.DownSingular}} struct {
	Conn *sqlx.DB
}

// New{{$alias.DownSingular}}Store returns a new store
func New{{$alias.DownSingular}}Store(conn *sqlx.DB) *{{$alias.DownSingular}} {
	os := &{{$alias.DownSingular}}{conn}
	return os
}

// All {{$alias.DownSingular}}
func (s *{{$alias.DownSingular}}) All() (db.{{$alias.DownSingular}}Slice, error) {
	return db.{{$alias.UpPlural}}().All(s.Conn)
}

// Get {{$alias.DownSingular}}
func (s *{{$alias.DownSingular}}) Get(id uuid.UUID) (*db.{{$alias.DownSingular}}, error) {
	return db.Find{{$alias.DownSingular}}(s.Conn, id.String())
}

// GetMany {{$alias.DownSingular}}
func (s *{{$alias.DownSingular}}) GetMany(keys []string) (db.{{$alias.DownSingular}}Slice, []error) {
	if len(keys) == 0 {
		return nil, []error{fmt.Errorf("no keys provided")}
	}
	args := []interface{}{}
	for _, key := range keys {
		args = append(args, key)
	}
	records, err := db.{{$alias.UpPlural}}(qm.WhereIn("id in ?", args...)).All(s.Conn)
	if errors.Is(err, sql.ErrNoRows) {
		return []*db.{{$alias.DownSingular}}{}, nil
	}
	if err != nil {
		return nil, []error{err}
	}

	result := []*db.{{$alias.DownSingular}}{}
	for _, key := range keys {
		for _, record := range records {
			if record.ID == key {
				result = append(result, record)
				break
			}
		}
	}
	return result, nil
}

// Insert {{$alias.DownSingular}}
func (s *{{$alias.DownSingular}}) Insert(record *db.{{$alias.DownSingular}}, txes ...*sql.Tx) (*db.{{$alias.DownSingular}}, error) {
	var err error

	handleTransactions(s.Conn, func(tx *sql.Tx) error {
		return record.Insert(tx, boil.Infer())
	}, txes...)

	err = record.Reload(s.Conn)
	if err != nil {
		return nil, err
	}
	return record, err
}

// Update {{$alias.DownSingular}}
func (s *{{$alias.DownSingular}}) Update(record *db.{{$alias.DownSingular}}, txes ...*sql.Tx) (*db.{{$alias.DownSingular}}, error) {
	_, err := record.Update(s.Conn, boil.Infer())
	if err != nil {
		return nil, err
	}
	return record, nil
}
