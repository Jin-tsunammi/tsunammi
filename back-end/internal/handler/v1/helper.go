package v1

import (
	"mm/pkg/apperrors"
	"strconv"
)

func parsePaginationParams(page, pageSize string) (parsedPage, parsedPageSize int, err error) {
	if page != "" {
		parsedPage, err = strconv.Atoi(page)
		if err != nil {
			return 0, 0, apperrors.BadRequest("page is not valid", err)
		}

		if parsedPage == 0 {
			return 0, 0, apperrors.BadRequest("page must be greater than 0")
		}
	}

	if pageSize != "" {
		parsedPageSize, err = strconv.Atoi(pageSize)
		if err != nil {
			return 0, 0, apperrors.BadRequest("pageSize is not valid", err)
		}

		if parsedPageSize > 100 || parsedPageSize == 0 {
			return 0, 0, apperrors.BadRequest("pageSize must be between 1 and 100")
		}
	}

	if parsedPage < 0 || parsedPageSize < 0 {
		return 0, 0, apperrors.BadRequest("page or pageSize must be positive")
	}

	return parsedPage, parsedPageSize, nil
}
