import * as slack from '@pulumi/slack'

start()

async function start() {
  var user1 = await slack.lookupUser({ email: "user1@email.com" })
  var user2 = await slack.lookupUser({ email: "user2@email.com" })

  const project = new slack.Conversation('channel1', {
    name: 'channel1',
    isPrivate: true,
    members: [user1.user.id, user2.user.id]
  })
}