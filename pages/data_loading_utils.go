package pages

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/house-emoji/bhap"
	"google.golang.org/appengine/datastore"
)

// bhapFromURLVars looks in the provided URL variables for a BHAP identifier
// and loads the BHAP from it.
//
// In this case, URL variables refer to what mux.Vars(r) returns.
func bhapFromURLVars(ctx context.Context, vars map[string]string) (bhap.BHAP, *datastore.Key, error) {
	if idStr, ok := vars["id"]; ok {
		// The BHAP is being identified by its regular ID
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return bhap.BHAP{}, nil, fmt.Errorf("converting BHAP ID to int: %v", err)
		}

		return bhap.ByID(ctx, id)
	} else if draftID, ok := vars["draftID"]; ok {
		// The BHAP is being identified by its draft ID
		return bhap.ByDraftID(ctx, draftID)
	} else {
		// Invalid request
		return bhap.BHAP{}, nil, errors.New("no provided BHAP identifier")
	}
}
