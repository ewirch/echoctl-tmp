- use mqtt 5
- expose echoctl settings in add-on
- post device/sensor configuration to mqtt
  - read sensor configuration
  - post on start
  - check for modification (mqtt server restart), and re-post
- DEBUG cmd option
- mqtt write channel
  - configure writable sensors
  - configure sensors to listen to (alternative: listen to all writable sensors)
  - listen to a key
  - post to canbus
