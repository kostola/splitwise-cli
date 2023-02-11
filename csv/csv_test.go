package csv_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kostola/splitwise-cli/csv"
)

func TestRead(t *testing.T) {
	expectedEntries := loadExpectedEntries(t)

	tests := []struct {
		name          string
		expectedError string
	}{
		{
			name: "ok_standard",
		},
		{
			name: "ok_scrambled",
		},
		{
			name:          "err_missing_header",
			expectedError: "missing \"users__0__paid_share\" header",
		},
		{
			name:          "err_wrong_number_of_fields",
			expectedError: "record on line 4: wrong number of fields",
		},
		{
			name:          "err_wrong_field_type",
			expectedError: "strconv.ParseFloat: parsing \"not a float\": invalid syntax: error on row 3",
		},
		{
			name:          "err_total_paid_share",
			expectedError: "total paid share is not 100 (87.500000): error on row 1",
		},
		{
			name:          "err_total_owed_share",
			expectedError: "total owed share is not 100 (81.250000): error on row 5",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entries, err := csv.Read(fmt.Sprintf("testdata/%s.csv", test.name))
			if len(test.expectedError) > 0 {
				require.ErrorContains(t, err, test.expectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, expectedEntries, entries)
			}
		})
	}
}

func loadExpectedEntries(t *testing.T) []csv.Entry {
	expectedBytes, err := os.ReadFile("testdata/expected.json")
	require.NoError(t, err)

	var expectedEntries []csv.Entry
	err = json.Unmarshal(expectedBytes, &expectedEntries)
	require.NoError(t, err)

	return expectedEntries
}
