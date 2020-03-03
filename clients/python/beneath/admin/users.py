from datetime import datetime

from beneath.connection import Connection
from beneath.utils import format_graphql_time


class Users:

  def __init__(self, conn: Connection):
    self.conn = conn

  async def get_me(self):
    """
      Returns info about the authenticated user.
      Returns None if authenticated with a project secret.
    """
    result = await self.conn.query_control(
      variables={},
      query="""
        query Me {
          me {
            userID
            user {
              username
              name
              bio
              photoURL
              createdOn
            }
            email
            organization {
              organizationID
              name
              personal
            }
            updatedOn
          }
        }
      """
    )
    me = result['me']
    if me is None:
      raise Exception("Cannot call get_me when authenticated with a service key")
    return me

  async def get_by_id(self, user_id):
    """
      Returns info about the user with `user_id`.
    """
    result = await self.conn.query_control(
      variables={
        'userID': user_id,
      },
      query="""
        query User($userID: UUID!) {
          user(
            userID: $userID
          ) {
            userID
            username
            name
            bio
            photoURL
            createdOn
            projects {
              projectID
              name
              displayName
              site
              description
              photoURL
              public
              createdOn
              updatedOn
              streams {
                name
              }
            }
          }
        }
      """
    )
    return result['user']

  async def get_by_username(self, username):
    """
      Returns info about the user with `username`.
    """
    result = await self.conn.query_control(
      variables={
        'username': username,
      },
      query="""
        query UserByUsername($username: String!) {
          userByUsername(
            username: $username
          ) {
            userID
            username
            name
            bio
            photoURL
            createdOn
            projects {
              projectID
              name
              displayName
              site
              description
              photoURL
              public
              createdOn
              updatedOn
              streams {
                name
              }
            }
          }
        }
      """
    )
    return result['userByUsername']

  async def get_usage(self, user_id, period=None, from_time=None, until=None):
    today = datetime.today()
    if (period is None) or (period == 'M'):
      default_time = datetime(today.year, today.month, 1)
    elif period == 'H':
      default_time = datetime(today.year, today.month, today.day, today.hour)

    result = await self.conn.query_control(
      variables={
        'userID': user_id,
        'period': period if period else 'M',
        'from': format_graphql_time(from_time) if from_time else format_graphql_time(default_time),
        'until': format_graphql_time(until) if until else None
      },
      query="""
        query GetUserMetrics($userID: UUID!, $period: String!, $from: Time!, $until: Time) {
          getUserMetrics(userID: $userID, period: $period, from: $from, until: $until) {
            entityID
            period
            time
            readOps
            readBytes
            readRecords
            writeOps
            writeBytes
            writeRecords
          }
        }
      """
    )
    return result['getUserMetrics']
