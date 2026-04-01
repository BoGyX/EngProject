#!/bin/sh
# Добавляем securityDefinitions в swagger.json

SWAGGER_FILE="docs/swagger.json"

# Проверяем, есть ли уже securityDefinitions
if grep -q "securityDefinitions" "$SWAGGER_FILE"; then
    echo "securityDefinitions already exists"
    exit 0
fi

# Добавляем securityDefinitions перед последней закрывающей скобкой
sed -i '$ d' "$SWAGGER_FILE"
cat >> "$SWAGGER_FILE" << 'EOF'
    },
    "securityDefinitions": {
        "BearerAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header",
            "description": "Type 'Bearer' followed by a space and JWT token."
        }
    }
}
EOF

echo "securityDefinitions added successfully"
