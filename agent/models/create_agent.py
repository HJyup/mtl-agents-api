from agents import Agent, set_default_openai_key
from agent.models import calendar_agent, gateway_agent

from agent.protos import config_pb2


def create_agent_for_user(config: config_pb2.GetConfigurationResponse) -> Agent:
    if not config:
        raise ValueError("Configuration is required to create an agent")

    if config.open_ai_key:
        set_default_openai_key(config.open_ai_key)

    handoffs = []

    if len(config.calendar.google_api_key) > 0:
        handoffs.append(calendar_agent)

    gateway_agent.handoffs = handoffs

    return gateway_agent