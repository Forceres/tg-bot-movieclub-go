package kinopoisk

import "regexp"

var kinopoiskURL = regexp.MustCompile(`https://www.kinopoisk.ru/film/(\d+)/`)

func ParseIDsOrRefs(rawString string) []string {
	matches := kinopoiskURL.FindAllStringSubmatch(rawString, -1)
	var ids []string
	for _, match := range matches {
		if len(match) > 1 && !contains(&ids, match[1]) {
			ids = append(ids, match[1])
		}
	}
	return ids
}

func contains(slice *[]string, item string) bool {
	for _, s := range *slice {
		if s == item {
			return true
		}
	}
	return false
}
