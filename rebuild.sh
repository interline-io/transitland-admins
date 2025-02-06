#!/bin/bash
go install cmd/transitland-admins
transitland-admins -idkey tzid zipgeojson https://github.com/evansiroky/timezone-boundary-builder/releases/download/2024a/timezones-now.geojson.zip timezones/timezones.polyline
transitland-admins -idkey name -include name,iso_a2,iso_3166_2,admin shapefile https://naciscdn.org/naturalearth/10m/cultural/ne_10m_admin_1_states_provinces.zip admins/admins.polyline
