package slice

import "strconv"

func GetByIndex(slice [][]string, index int64) *[]string {
	for _, item := range slice {
		itemIndex, err := strconv.ParseInt(item[0], 10, 64)
		if err != nil {
			continue
		}
		if itemIndex == index {
			return &item
		}
	}
	return nil
}
