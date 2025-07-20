ENV_FILE := .env
API_CONFIG_OUTPUT := ./api/config/config.yaml
API_CONFIG_TEMPLATE := api_template.txt
LISTING_CONFIG_OUTPUT := ./listing_service/.env
LISTING_CONFIG_TEMPLATE := listing_template.txt
USER_CONFIG_OUTPUT := ./user_service/.env
USER_CONFIG_TEMPLATE := user_template.txt
SESSION_CONFIG_OUTPUT := ./session_service/.env
SESSION_CONFIG_TEMPLATE := session_template.txt

.PHONY: all generate_api generate_listing generate_user generate_session

generate_api:
	@echo "Generating $(API_CONFIG_OUTPUT) from $(API_CONFIG_TEMPLATE)..."
	@export $$(cat $(ENV_FILE) | sed 's/ *= */=/' | grep -v '^#') && \
	envsubst < $(API_CONFIG_TEMPLATE) > $(API_CONFIG_OUTPUT)
	@echo "Done: $(API_CONFIG_OUTPUT) created."

generate_listing:
	@echo "Generating $(LISTING_CONFIG_OUTPUT) from $(LISTING_CONFIG_TEMPLATE)..."
	@export $$(cat $(ENV_FILE) | sed 's/ *= */=/' | grep -v '^#') && \
	envsubst < $(LISTING_CONFIG_TEMPLATE) > $(LISTING_CONFIG_OUTPUT)
	@echo "Done: $(LISTING_CONFIG_OUTPUT) created."

generate_user:
	@echo "Generating $(USER_CONFIG_OUTPUT) from $(USER_CONFIG_TEMPLATE)..."
	@export $$(cat $(ENV_FILE) | sed 's/ *= */=/' | grep -v '^#') && \
	envsubst < $(USER_CONFIG_TEMPLATE) > $(USER_CONFIG_OUTPUT)
	@echo "Done: $(USER_CONFIG_OUTPUT) created."

generate_session:
	@echo "Generating $(SESSION_CONFIG_OUTPUT) from $(SESSION_CONFIG_TEMPLATE)..."
	@export $$(cat $(ENV_FILE) | sed 's/ *= */=/' | grep -v '^#') && \
	envsubst < $(SESSION_CONFIG_TEMPLATE) > $(SESSION_CONFIG_OUTPUT)
	@echo "Done: $(SESSION_CONFIG_OUTPUT) created."

all: generate_api generate_listing generate_user generate_session

