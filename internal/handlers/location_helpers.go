package handlers

import (
	"strings"

	"kariakoo/inventory/internal/models"
)

func storeLocationsOnly(locations []*models.BusinessLocation) []*models.BusinessLocation {
	filtered := make([]*models.BusinessLocation, 0, len(locations))
	for _, loc := range locations {
		if strings.EqualFold(loc.LocationType, "store") {
			filtered = append(filtered, loc)
		}
	}

	if len(filtered) == 0 {
		return locations
	}

	return filtered
}

func locationByID(locations []*models.BusinessLocation, id int) *models.BusinessLocation {
	for _, loc := range locations {
		if loc.ID == id {
			return loc
		}
	}
	return nil
}

func firstStoreLocation(locations []*models.BusinessLocation) *models.BusinessLocation {
	for _, loc := range locations {
		if strings.EqualFold(loc.LocationType, "store") {
			return loc
		}
	}

	if len(locations) > 0 {
		return locations[0]
	}

	return nil
}
