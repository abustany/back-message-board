package poststore

import (
	"encoding/csv"
	"io"
	"time"

	"github.com/pkg/errors"

	"github.com/abustany/back-message-board/pkg/types"
)

// LoadFromCSV loads the contents of a CSV file into a Store, and return the
// number of posts inserted.
//
// The CSV records must have 5 columns: id, name, email, text, created (in RFC3339 format).
func LoadFromCSV(store Store, data io.Reader, hasHeader bool) (uint, error) {
	reader := csv.NewReader(data)

	// Id, name, email, text, created
	reader.FieldsPerRecord = 5
	reader.ReuseRecord = true

	counter := uint(0)

	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			return counter, errors.Wrap(err, "Error while decoding CSV file")
		}

		counter++

		if hasHeader && counter == 1 {
			hasHeader = false
			counter = 0
			continue
		}

		created, err := time.Parse(time.RFC3339, record[4])

		if err != nil {
			return counter, errors.Wrapf(err, "Error while parsing creation date of record %d", counter)
		}

		post := types.Post{
			ID:      record[0],
			Author:  record[1],
			Email:   record[2],
			Message: record[3],
			Created: created,
		}

		if err := store.Add(post); err != nil {
			return counter, errors.Wrapf(err, "Error while inserting post for record %d", counter)
		}
	}

	return counter, nil
}
