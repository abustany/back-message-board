package poststore_test

import (
	"strings"
	"testing"

	"github.com/abustany/back-message-board/pkg/poststore"
)

func TestLoadFromCSV(t *testing.T) {
	const header = "id,name,email,text,created\n"
	const record = "ID,John,john@domain.com,Hello world,2017-12-14T06:20:33-08:00\n"
	const recordWrongNFields = "Not,the,right,number\n"
	const recordIncorrectTime = "ID,John,john@domain.com,Hello world,yesterday around 3 o clock\n"

	testData := []struct {
		name      string
		data      string
		hasHeader bool
		nRecords  int
	}{
		{
			"Record with no header",
			record,
			false,
			1,
		},
		{
			"Record with a header",
			header + record,
			true,
			1,
		},
		{
			"Empty file with no header",
			"",
			false,
			0,
		},
		{
			"Empty file with a header",
			header,
			true,
			0,
		},
		{
			"Wrong number of fields",
			header + recordWrongNFields,
			true,
			-1,
		},
		{
			"Incorrect time",
			header + recordIncorrectTime,
			true,
			-1,
		},
	}

	for _, d := range testData {
		store, err := poststore.NewMemoryPostStore()

		if err != nil {
			t.Fatalf("Error while creating store: %s", err)
		}

		n, err := poststore.LoadFromCSV(store, strings.NewReader(d.data), d.hasHeader)

		if d.nRecords >= 0 {
			if err != nil {
				t.Errorf("LoadFromCSV returned an error for case %s: %s", d.name, err)
			}

			if n != uint(d.nRecords) {
				t.Errorf("LoadFromCSV returned an incorrect record count for case %s: got %d, expected %d", d.name, n, d.nRecords)
			}
		} else {
			if err == nil {
				t.Errorf("LoadFromCSV returned no error for case %s", d.name)
			}
		}
	}
}
