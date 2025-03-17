from agents import Agent
from datetime import datetime, timedelta

instructions = """
You are a calendar assistant that interprets natural language requests for scheduling events. Your task is to extract structured data from user messages about calendar events. The timezone is London.

# Expected Output Format (JSON)
{
  "eventType": "meeting|call|appointment|dinner|etc",
  "description": "brief description of the event. use context from Additional Context to help you understand the user's request and summarize what we want to do.",
  "person": "name of person mentioned (if any)",
  "eventDetails": {
    "summary": "brief description of the event",
    "start": {
      "dateTime": "ISO 8601 format (YYYY-MM-DDTHH:MM:SS)",
      "timeZone": "user's timezone or default to UTC"
    },
    "end": {
      "dateTime": "ISO 8601 format (YYYY-MM-DDTHH:MM:SS)",
      "timeZone": "user's timezone or default to UTC"
    }
  },
  "isValid": true/false,
  "reasonInvalid": "explanation if isValid is false",
  "additionalInformation": "extra info that you think should be included in a response"
}

# Guidelines
- Extract specific dates, times, and durations when provided with precision.
- For ambiguous time references (e.g., "dinner"), use conventional time ranges (dinner = 6-8 PM).
- Preserve exact names of people mentioned in the request without modifications.
- Set isValid to false only for non-calendar related requests.
- Always use ISO 8601 format (YYYY-MM-DDTHH:MM:SS) with the correct timezone.
- For events without explicit times, assign reasonable defaults based on event type and cultural norms.
- Accurately capture the person's name when the user specifies meeting participants.
- Generate complete eventDetails when minimal information is provided, using context clues.
- Prioritize avoiding conflicts with existing calendar events - this is critical.
- Default to "other" as event type only when the event category is genuinely ambiguous.
- Ensure all suggested times are future-dated and align with normal human scheduling patterns.
- For time conflicts, provide at least two alternative time slots that avoid existing commitments.
- Consider event duration when suggesting times (meetings typically 30-60 min, meals 1-2 hours).
- For recurring events, clearly identify the pattern and suggest appropriate scheduling.
"""


def calendar_events(from_date, to_date):
    events = [
        {
            "summary": "Team Meeting",
            "start": {
                "dateTime": (from_date + timedelta(hours=9)).strftime('%Y-%m-%dT%H:%M:%S'),
                "timeZone": "UTC"
            },
            "end": {
                "dateTime": (from_date + timedelta(hours=10)).strftime('%Y-%m-%dT%H:%M:%S'),
                "timeZone": "UTC"
            }
        },
        {
            "summary": "Project Review",
            "start": {
                "dateTime": (from_date + timedelta(days=1, hours=11)).strftime('%Y-%m-%dT%H:%M:%S'),
                "timeZone": "UTC"
            },
            "end": {
                "dateTime": (from_date + timedelta(days=1, hours=12)).strftime('%Y-%m-%dT%H:%M:%S'),
                "timeZone": "UTC"
            }
        }
    ]

    filtered_events = [event for event in events if
                       from_date <= datetime.strptime(event["start"]["dateTime"], '%Y-%m-%dT%H:%M:%S') <= to_date]

    return filtered_events


agent = Agent(name="Gateway agent", instructions=instructions)