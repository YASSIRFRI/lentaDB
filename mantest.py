import requests
import random
import string
import time

BASE_URL = "http://localhost:8080"
key_values = {}
outputfile = open("output.txt", "w")

def generate_random_key():
    key_length = random.randint(5, 10)
    return ''.join(random.choice(string.ascii_letters) for _ in range(key_length))

def generate_random_value():
    value_length = random.randint(5, 15)
    return ''.join(random.choice(string.ascii_letters + string.digits) for _ in range(value_length))

def send_set_request(key,value):
    #key_values[key] = value
    response = requests.post(f"{BASE_URL}/set", data={"key": key, "value": value})
    outputfile.write(f"SET Response: {response.status_code} - {response.text}\n")

def send_get_request(key):
    start_time = time.time()
    response = requests.get(f"{BASE_URL}/get?key={key}")
    elapsed_time = time.time() - start_time
    outputfile.write(f"GET Response: {response.status_code} - {response.text}, Elapsed Time: {elapsed_time:.6f} seconds\n")

def send_delete_request(key):
    response = requests.delete(f"{BASE_URL}/del?key={key}")
    outputfile.write(f"DELETE Response: {response.status_code} - {response.text}\n")

# Set 1000 random keys
for i in range(1000):
    send_set_request(i,i)

## Query each key
for i in range(1000):
    send_get_request(i)

for i in range(1000):
    send_delete_request(i)

#for key in key_values.keys():
    #send_get_request(key)

#send_set_request("yassir","fri")
#get_response = send_get_request("yassir")
#delete_response = send_delete_request("yassir")
#get_response = send_get_request("yassir")