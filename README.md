#Go base clean

##Run

####rsa_keys creations:
1. Create a folder "rsa_keys" in the root of the project.
2. Create file "private_key.pem" and "public_key.pem".
3. Generate keys at "https://travistidwell.com/jsencrypt/demo/".
4. Add keys to relevant files.

First start use:

    make start

Then use:

    make run

Returns current aggregated weather for specified city

curl -X GET "http://localhost:8080/api/v1/weather/current?city=London" \
-H "Accept: application/json"


Returns aggregated forecast data with validated 'days' parameter

curl -X GET "http://localhost:8080/api/v1/weather/forecast?city=London&days=3" \
-H "Accept: application/json"


Returns service health status and last successful API fetch times

curl -X GET "http://localhost:8080/api/v1/health" \
-H "Accept: application/json"