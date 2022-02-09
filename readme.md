# Egnyte Elasticsearch Pipeline

The purpose of this application is to provide useful monitoring data to an elasticsearch instance with little trouble.

## Testing this Application

If you would like to test the application to see what data it generates, please feel free to do so by following the steps below.
The development environment will be deployed locally using Docker Compose to deploy the serivce, elasticsearch, and kibana.

1. Clone the repo locally to a machine with Docker and Docker Compose installed.
2. Copy the config.example.json file to config.json.
3. Modify the parameters within the new config file to match your Egnyte step. <br/>
   <strong>Note</strong>: you can ignore the elastic configurations for development.
4. Run the following command to build the docker container.<br>
   TODO: Add command below.
5. Run docker-compose -d up to deploy development environment. Will host a local Kibana and Elasticsearch instance for the Engyte ES Service to push data to.
6. Access the kibana dashboard at http://localhost:5601/app/home#/

## MVP Task Lists

- [ ] Publish events data to elasticsearch.
- [ ] Allow for configuration of remote elasticsearch instances.
- [ ] Track application data in state for users and cursor.
- [ ] Correlate user names with actor IDs prior to sending event data to elastic.

## Future Application Goals

Coming Soon.

## Contributions

Feel free to contribute to this project.
This project is currently being developed internally by Applied Network Solutions, Inc.
Once phase one of roadmap is complete, we will be using the product internally generate data for our elastic SIEM.
