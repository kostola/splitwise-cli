package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"math/big"
	"os"
	"strconv"
)

const (
	groupId       string = "group_id"
	date          string = "date"
	cost          string = "cost"
	currency      string = "currency"
	categoryId    string = "category_id"
	description   string = "description"
	details       string = "details"
	userId        string = "users__%d__user_id"
	userPaidShare string = "users__%d__paid_share"
	userOwedShare string = "users__%d__owed_share"
)

type Indexes struct {
	groupId     int
	date        int
	cost        int
	currency    int
	categoryId  int
	description int
	details     int
	users       []UserIndexes
}

type UserIndexes struct {
	userId    int
	paidShare int
	owedShare int
}

type Entry struct {
	GroupId     string
	Date        string
	Cost        float64
	Currency    string
	CategoryId  string
	Description string
	Details     string
	Users       []UserEntry
}

type UserEntry struct {
	UserId    string
	PaidShare float64
	OwedShare float64
}

func Read(filePath string) ([]Entry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	result, err := read(csv.NewReader(file))
	if err != nil {
		return nil, err
	}
	return result, nil
}

func read(r *csv.Reader) ([]Entry, error) {
	idxs, err := readIndexes(r)
	if err != nil {
		return nil, err
	}

	var entries []Entry

	for row := 1; true; row++ {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		entry, err := entryFromRecord(record, &idxs)
		if err != nil {
			return nil, fmt.Errorf("%w: error on row %d", err, row)
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func entryFromRecord(record []string, idxs *Indexes) (Entry, error) {
	entry := Entry{
		GroupId:     record[idxs.groupId-1],
		Date:        record[idxs.date-1],
		Currency:    record[idxs.currency-1],
		CategoryId:  record[idxs.categoryId-1],
		Description: record[idxs.description-1],
		Details:     record[idxs.details-1],
	}

	cost, err := strconv.ParseFloat(record[idxs.cost-1], 64)
	if err != nil {
		return Entry{}, err
	}
	entry.Cost = cost

	var totalPaidShare, totalOwedShare float64

	for _, u := range idxs.users {
		paidShare, err := strconv.ParseFloat(record[u.paidShare-1], 64)
		if err != nil {
			return Entry{}, err
		}

		owedShare, err := strconv.ParseFloat(record[u.owedShare-1], 64)
		if err != nil {
			return Entry{}, err
		}

		totalPaidShare += paidShare
		totalOwedShare += owedShare

		entry.Users = append(entry.Users, UserEntry{UserId: record[u.userId-1], PaidShare: paidShare, OwedShare: owedShare})
	}

	if big.NewFloat(100.0).Cmp(big.NewFloat(totalPaidShare)) != 0 {
		return Entry{}, fmt.Errorf("total paid share is not 100 (%f)", totalPaidShare)
	}
	if big.NewFloat(100.0).Cmp(big.NewFloat(totalOwedShare)) != 0 {
		return Entry{}, fmt.Errorf("total owed share is not 100 (%f)", totalOwedShare)
	}

	return entry, nil
}

func readIndexes(r *csv.Reader) (Indexes, error) {
	headers, err := r.Read()
	if err != nil {
		return Indexes{}, fmt.Errorf("%w: can't read header row", err)
	}
	cfi := Indexes{}

	for i, header := range headers {
		switch header {
		case groupId:
			cfi.groupId = i + 1
		case date:
			cfi.date = i + 1
		case cost:
			cfi.cost = i + 1
		case currency:
			cfi.currency = i + 1
		case categoryId:
			cfi.categoryId = i + 1
		case description:
			cfi.description = i + 1
		case details:
			cfi.details = i + 1
		}
	}

	for u := 0; u < 10; u++ {
		for i, header := range headers {
			if header == fmt.Sprintf(userId, u) {
				for len(cfi.users) <= u {
					cfi.users = append(cfi.users, UserIndexes{})
				}
				cfi.users[u].userId = i + 1
			} else if header == fmt.Sprintf(userPaidShare, u) {
				for len(cfi.users) <= u {
					cfi.users = append(cfi.users, UserIndexes{})
				}
				cfi.users[u].paidShare = i + 1
			} else if header == fmt.Sprintf(userOwedShare, u) {
				for len(cfi.users) <= u {
					cfi.users = append(cfi.users, UserIndexes{})
				}
				cfi.users[u].owedShare = i + 1
			}
		}
	}

	if err = validateIndexes(&cfi); err != nil {
		return Indexes{}, err
	}

	return cfi, nil
}

func validateIndexes(i *Indexes) error {
	if i.groupId < 1 {
		return fmt.Errorf("missing \"%s\" header", groupId)
	}
	if i.date < 1 {
		return fmt.Errorf("missing \"%s\" header", date)
	}
	if i.cost < 1 {
		return fmt.Errorf("missing \"%s\" header", cost)
	}
	if i.currency < 1 {
		return fmt.Errorf("missing \"%s\" header", currency)
	}
	if i.categoryId < 1 {
		return fmt.Errorf("missing \"%s\" header", categoryId)
	}
	if i.description < 1 {
		return fmt.Errorf("missing \"%s\" header", description)
	}
	if i.details < 1 {
		return fmt.Errorf("missing \"%s\" header", details)
	}
	return validateUserIndexesSlice(i.users)
}

func validateUserIndexesSlice(indexes []UserIndexes) error {
	for i, item := range indexes {
		if err := validateUserIndexes(i, &item); err != nil {
			return err
		}
	}
	return nil
}

func validateUserIndexes(i int, index *UserIndexes) error {
	if index.userId < 1 {
		return fmt.Errorf("missing \"%s\" header", fmt.Sprintf(userId, i))
	}
	if index.paidShare < 1 {
		return fmt.Errorf("missing \"%s\" header", fmt.Sprintf(userPaidShare, i))
	}
	if index.owedShare < 1 {
		return fmt.Errorf("missing \"%s\" header", fmt.Sprintf(userOwedShare, i))
	}
	return nil
}
