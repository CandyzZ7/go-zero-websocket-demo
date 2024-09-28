.PHONY: api
# generate api proto
api:
	goctl api go -api ./api/app.api -dir . -style go_zero -home=./tpl