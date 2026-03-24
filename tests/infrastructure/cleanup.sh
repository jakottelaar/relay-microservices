#!/bin/bash

CONTAINER_NAME="test-relay-postgres"
DB_NAMES=("test_relay_auth" "test_relay_users" "test_relay_guilds")
DB_USER="test_relay"

# Iterate through each database
for DB_NAME in "${DB_NAMES[@]}"; do
    echo "Truncating tables in $DB_NAME..."
    
    # Generate TRUNCATE statements for all tables
    TRUNCATE_SQL=$(docker exec -i $CONTAINER_NAME psql -U $DB_USER -d $DB_NAME -t -c \
    "SELECT 'TRUNCATE TABLE \"' || tablename || '\" CASCADE;' FROM pg_tables WHERE schemaname='public';")
    
    # Execute the TRUNCATE statements
    docker exec -i $CONTAINER_NAME psql -U $DB_USER -d $DB_NAME -c "$TRUNCATE_SQL"
done

echo "Done!"