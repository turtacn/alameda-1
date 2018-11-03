**Architecture of Alameda**

Alameda aims to make the resources orchestrator for Kubernetes easy and intelligent. It plays the role as the brain of Kubernertes resources orchestrator, and generally includes three modules, the data collector module, the AI module, and the execution module. 

* Data Collector: Retreiving metrics from Prometheous and passing data to the AI Module
* AI Service: Generating prediction results based on multiple machine learning algorithms. Also generating resource configuration plans.
* Execution Service: Alameda doenes't provide any execution service. Instead, any third parties can execute the resource configuration plans on their own.
