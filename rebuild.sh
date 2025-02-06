#!/bin/bash
transitland polylines-create --idkey tzid zipgeojson "https://github.com/evansiroky/timezone-boundary-builder/releases/download/2024a/timezones-now.geojson.zip" "timezones/timezones.polyline"
transitland polylines-create --idkey name -k name -k iso_a2 -k iso_3166_2 -k admin shapefile "https://naciscdn.org/naturalearth/10m/cultural/ne_10m_admin_1_states_provinces.zip" "admins/admins.polyline"
