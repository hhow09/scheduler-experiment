build_images:
	cd services/report-apis && docker build -t report-apis . && cd ../..
	cd services/report-collector && docker build -t report-collector . && cd ../..

start_all:
	kubectl apply -f apis.yml
	kubectl apply -f jobs.yml

remove_all:
	kubectl delete service report-apis-service
	kubectl delete deployment report-apis
	kubectl delete cronjob report-collector
