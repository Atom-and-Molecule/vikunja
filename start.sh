#!/bin/bash
set -e

# Parse DATABASE_URL if available
if [ -n "$DATABASE_URL" ]; then
    echo "Configuring database connection from DATABASE_URL..."
    # URL format: postgres://user:password@host:port/database
    
    # 1. Strip protocol (postgres:// or postgresql://)
    URL_WITHOUT_PROTO="${DATABASE_URL#*://}"
    
    # 2. Extract user:password part (everything before the last @)
    USER_PASS="${URL_WITHOUT_PROTO%%@*}"
    export VIKUNJA_DATABASE_USER="${USER_PASS%%:*}"
    export VIKUNJA_DATABASE_PASSWORD="${USER_PASS#*:}"
    
    # 3. Extract host:port/database part (everything after the last @)
    HOST_PORT_DB="${URL_WITHOUT_PROTO##*@}"
    
    # 4. Extract database name (everything after the slash)
    DB_NAME_WITH_QUERY="${HOST_PORT_DB#*/}"
    # Remove any query parameters like ?sslmode=disable
    export VIKUNJA_DATABASE_DATABASE="${DB_NAME_WITH_QUERY%%\?*}"
    
    # 5. Extract host and port
    HOST_PORT="${HOST_PORT_DB%%/*}"
    export VIKUNJA_DATABASE_HOST="$HOST_PORT"
    
    export VIKUNJA_DATABASE_TYPE="postgres"
    
    # If the database URL has a query string, look for sslmode
    if [[ "$DB_NAME_WITH_QUERY" == *\?* ]]; then
        QUERY_STRING="${DB_NAME_WITH_QUERY#*\?}"
        # Split query parameters by &
        IFS='&' read -ra PARAMS <<< "$QUERY_STRING"
        for param in "${PARAMS[@]}"; do
            key="${param%%=*}"
            value="${param#*=}"
            if [ "$key" = "sslmode" ]; then
                export VIKUNJA_DATABASE_SSLMODE="$value"
            fi
        done
    fi
    
    # If SSL mode was not set by the URL query parameter, default to disable
    if [ -z "$VIKUNJA_DATABASE_SSLMODE" ]; then
        export VIKUNJA_DATABASE_SSLMODE="disable"
    fi

    echo "Database configured: type=postgres, host=$VIKUNJA_DATABASE_HOST, user=$VIKUNJA_DATABASE_USER, database=$VIKUNJA_DATABASE_DATABASE, sslmode=$VIKUNJA_DATABASE_SSLMODE"
fi

# Wait for database to be ready before starting the service
if [ -n "$VIKUNJA_DATABASE_HOST" ]; then
    DB_HOST_ONLY="${VIKUNJA_DATABASE_HOST%%:*}"
    DB_PORT_ONLY="${VIKUNJA_DATABASE_HOST#*:}"
    if [ "$DB_PORT_ONLY" = "$VIKUNJA_DATABASE_HOST" ]; then
        DB_PORT_ONLY=5432
    fi

    echo "Waiting for database to accept connections at $DB_HOST_ONLY:$DB_PORT_ONLY..."
    for i in {1..30}; do
        if bash -c "exec 3<>/dev/tcp/$DB_HOST_ONLY/$DB_PORT_ONLY" 2>/dev/null; then
            echo "Database is ready!"
            break
        fi
        echo "Database not ready yet, retrying in 2 seconds... ($i/30)"
        sleep 2
    done
fi

# Map PORT variable to VIKUNJA_SERVICE_INTERFACE
if [ -n "$PORT" ]; then
    export VIKUNJA_SERVICE_INTERFACE=":$PORT"
else
    export VIKUNJA_SERVICE_INTERFACE=":3456"
fi

# Set a random service secret if not already set (used for JWT sign/crypt)
if [ -z "$VIKUNJA_SERVICE_SECRET" ]; then
    echo "VIKUNJA_SERVICE_SECRET is not set. Generating a random one..."
    export VIKUNJA_SERVICE_SECRET=$(openssl rand -hex 32)
fi

# If PUBLICURL is not configured, set it using static domain if available
if [ -z "$VIKUNJA_SERVICE_PUBLICURL" ]; then
    if [ -n "$RAILWAY_STATIC_URL" ]; then
        export VIKUNJA_SERVICE_PUBLICURL="https://$RAILWAY_STATIC_URL"
    elif [ -n "$RAILWAY_PUBLIC_DOMAIN" ]; then
        export VIKUNJA_SERVICE_PUBLICURL="https://$RAILWAY_PUBLIC_DOMAIN"
    else
        export VIKUNJA_SERVICE_PUBLICURL="http://localhost:${PORT:-3456}"
    fi
    echo "Set VIKUNJA_SERVICE_PUBLICURL to $VIKUNJA_SERVICE_PUBLICURL"
fi

echo "Starting Vikunja..."
exec ./vikunja
