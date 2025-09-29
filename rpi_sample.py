# import json, time, random
# import paho.mqtt.client as mqtt

# MQTT_HOST = "YOUR_BROKER_HOST"   # e.g., "localhost" or broker IP
# MQTT_PORT = 1883
# MQTT_USER = None                  # or "user"
# MQTT_PASS = None                  # or "pass"

# PI_ID = "pi_demo_001"             # must already exist on server
# DEVICE_ID = 101
# TOPIC = f"sensors/{PI_ID}/{DEVICE_ID}/reading"

# client = mqtt.Client(client_id=f"pi-{PI_ID}-{DEVICE_ID}")
# if MQTT_USER:
#     client.username_pw_set(MQTT_USER, MQTT_PASS)

# client.connect(MQTT_HOST, MQTT_PORT, keepalive=30)
# client.loop_start()

# try:
#     while True:
#         payload = {
#             "value": round(20 + 10 * random.random(), 2),
#             "ts": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
#         }
#         client.publish(TOPIC, json.dumps(payload), qos=1)
#         time.sleep(1)
# except KeyboardInterrupt:
#     pass
# finally:
#     client.loop_stop()
#     client.disconnect()