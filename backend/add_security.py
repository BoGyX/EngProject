#!/usr/bin/env python3
import json

swagger_file = "docs/swagger.json"

with open(swagger_file, 'r') as f:
    data = json.load(f)

# Добавляем securityDefinitions если их нет
if 'securityDefinitions' not in data:
    data['securityDefinitions'] = {
        "BearerAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header",
            "description": "Type 'Bearer' followed by a space and JWT token."
        }
    }
    
    with open(swagger_file, 'w') as f:
        json.dump(data, f, indent=4)
    
    print("securityDefinitions added successfully")
else:
    print("securityDefinitions already exists")
