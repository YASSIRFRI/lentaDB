import requests
import random
import string
import time

BASE_URL = "http://localhost:8080"
key_values = {}

def generate_random_key():
    key_length = random.randint(5, 10)
    return ''.join(random.choice(string.ascii_letters) for _ in range(key_length))

def generate_random_value():
    value_length = random.randint(5, 15)
    return ''.join(random.choice(string.ascii_letters + string.digits) for _ in range(value_length))

def send_set_request():
    key = generate_random_key()
    value = generate_random_value()
    key_values[key] = value
    response = requests.post(f"{BASE_URL}/set", data={"key": key, "value": value})
    print(f"SET Response: {response.status_code} - {response.text}")

def send_get_request(key):
    start_time = time.time()
    response = requests.get(f"{BASE_URL}/get?key={key}")
    elapsed_time = time.time() - start_time
    print(f"GET Response: {response.status_code} - {response.text}, Elapsed Time: {elapsed_time:.6f} seconds")

def send_delete_request(key):
    response = requests.delete(f"{BASE_URL}/del?key={key}")
    print(f"DELETE Response: {response.status_code} - {response.text}")

# Set 1000 random keys
for _ in range(1000):
    send_set_request()

# Query each key
for key in key_values.keys():
    send_get_request(key)

# Delete each key
for key in key_values.keys():
    send_delete_request(key)

# Query each key after deletion
for key in key_values.keys():
    send_get_request(key)
