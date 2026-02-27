"""Unit tests for MQTT Bridge - uses unittest (no pytest required)"""
import sys
import os
import unittest
from unittest.mock import MagicMock


def _add_bridge_to_sys_path() -> None:
    """
    Ensure src/production/MQT.Bridge is on sys.path so we can import mqtt_bridge.

    This works whether tests are run from the repo root (recommended) or from
    within the src/production/MQT.Bridge subtree.
    """
    here = os.path.dirname(os.path.abspath(__file__))
    # MQT.Bridge/tests/unit -> MQT.Bridge/tests -> MQT.Bridge -> production -> src -> repo root
    bridge_dir = os.path.abspath(os.path.join(here, "..", ".."))
    if os.path.isdir(bridge_dir):
        sys.path.insert(0, bridge_dir)
        vendor = os.path.join(bridge_dir, "vendor")
        if os.path.exists(vendor):
            sys.path.insert(0, vendor)


_add_bridge_to_sys_path()

from mqtt_bridge import MQTTBridge, BridgeConfig  # noqa: E402


class TestBridgeConfig(unittest.TestCase):
    def test_config_values(self):
        config = BridgeConfig(
            external_broker="ext.example.com",
            external_port=1883,
            local_broker="local",
            local_port=1884,
            topic_filter="custom/#",
        )
        self.assertEqual(config.external_broker, "ext.example.com")
        self.assertEqual(config.local_port, 1884)
        self.assertEqual(config.topic_filter, "custom/#")

    def test_from_env_returns_config(self):
        config = BridgeConfig.from_env()
        self.assertIsNotNone(config.topic_filter)
        self.assertIsNotNone(config.external_broker)


class TestMQTTBridge(unittest.TestCase):
    def test_init_with_config(self):
        config = BridgeConfig(
            external_broker="ext",
            external_port=1883,
            local_broker="local",
            local_port=1883,
            topic_filter="sensors/#",
        )
        bridge = MQTTBridge(config=config)
        self.assertEqual(bridge.config, config)
        self.assertFalse(bridge.connected_to_external)
        self.assertFalse(bridge.connected_to_local)
        self.assertFalse(bridge.running)

    def test_on_external_message_forwards_when_local_connected(self):
        config = BridgeConfig(
            external_broker="ext",
            external_port=1883,
            local_broker="local",
            local_port=1883,
            topic_filter="sensors/#",
        )
        mock_local = MagicMock()
        mock_local.publish.return_value = MagicMock(rc=0)

        bridge = MQTTBridge(config=config, external_client=MagicMock(), local_client=mock_local)
        bridge.connected_to_local = True

        mock_msg = MagicMock()
        mock_msg.topic = "sensors/pi1/dev1/temp"
        mock_msg.payload = b'{"value": 25.5}'
        mock_msg.qos = 1
        mock_msg.retain = False

        bridge.on_external_message(None, None, mock_msg)

        mock_local.publish.assert_called_once_with(
            "sensors/pi1/dev1/temp",
            b'{"value": 25.5}',
            qos=1,
            retain=False,
        )

    def test_on_external_message_drops_when_local_disconnected(self):
        config = BridgeConfig(
            external_broker="ext",
            external_port=1883,
            local_broker="local",
            local_port=1883,
            topic_filter="sensors/#",
        )
        mock_local = MagicMock()

        bridge = MQTTBridge(config=config, external_client=MagicMock(), local_client=mock_local)
        bridge.connected_to_local = False

        mock_msg = MagicMock()
        mock_msg.topic = "sensors/pi1/dev1/temp"
        mock_msg.payload = b'{"value": 25.5}'

        bridge.on_external_message(None, None, mock_msg)

        mock_local.publish.assert_not_called()

    def test_create_clients_skipped_when_provided(self):
        config = BridgeConfig(
            external_broker="ext",
            external_port=1883,
            local_broker="local",
            local_port=1883,
            topic_filter="sensors/#",
        )
        ext_client = MagicMock()
        loc_client = MagicMock()
        bridge = MQTTBridge(config=config, external_client=ext_client, local_client=loc_client)

        bridge.create_clients()

        self.assertIs(bridge.external_client, ext_client)
        self.assertIs(bridge.local_client, loc_client)


if __name__ == "__main__":
    unittest.main()
