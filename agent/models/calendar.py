from googleapiclient.discovery import build
from google.oauth2 import service_account
from dotenv import load_dotenv
from agents import Agent, function_tool

load_dotenv()

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

@function_tool
def get_google_people_contacts():
    SCOPES = ['https://www.googleapis.com/auth/contacts.readonly']
    SERVICE_ACCOUNT_FILE = 'path/to/your/service_account.json'

    credentials = service_account.Credentials.from_service_account_file(
        SERVICE_ACCOUNT_FILE, scopes=SCOPES
    )

    service = build('people', 'v1', credentials=credentials)

    results = service.people().connections().list(
        resourceName='people/me',
        pageSize=100,
        personFields='names,emailAddresses'
    ).execute()

    connections = results.get('connections', [])
    return connections


@function_tool
def fetch_google_calendar_events(start_date, end_date):
    SCOPES = ['https://www.googleapis.com/auth/calendar.readonly']
    SERVICE_ACCOUNT_FILE = 'path/to/your/service_account.json'

    credentials = service_account.Credentials.from_service_account_file(
        SERVICE_ACCOUNT_FILE, scopes=SCOPES
    )

    service = build('calendar', 'v3', credentials=credentials)

    events_result = service.events().list(
        calendarId='primary',
        timeMin=start_date.isoformat() + 'Z',
        timeMax=end_date.isoformat() + 'Z',
        singleEvents=True,
        orderBy='startTime'
    ).execute()

    events = events_result.get('items', [])
    return events

_agent = Agent(name="Gateway agent", instructions=instructions, tools=[get_google_people_contacts, fetch_google_calendar_events])