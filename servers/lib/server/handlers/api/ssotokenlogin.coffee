{ daisy }  = require 'bongo'
{ argv }   = require 'optimist'
KONFIG     = require('koding-config-manager').load("main.#{argv.c}")
{ secret } = KONFIG.jwt
Jwt        = require 'jsonwebtoken'
hat        = require 'hat'

{ setSessionCookie } = require '../../helpers'

module.exports = ssoTokenLogin = (req, res, next) ->

  { JUser, JSession, JAccount } = (require '../../bongo').models

  { token } = req.query
  user      = null
  group     = null
  account   = null
  username  = null

  unless token
    return res.status(400).send 'token is required'

  queue = [

    ->
      validateToken token, (err, data) ->
        return res.status(err.statusCode).send(err.message)  if err
        { username, group } = data

        # making sure subdomain is same with group slug
        unless group in req.subdomains
          return res.status(400).send('invalid request')

        queue.next()

    ->
      # checking if user exists
      JAccount.one { 'profile.nickname' : username }, (err, account_) ->
        return res.status(500).send 'an error occurred'  if err
        return res.status(400).send 'invalid username!'  unless account_
        account = account_
        queue.next()

    ->
      # checking if user is a member of the group of api token
      client = { connection : { delegate : account } }
      account.checkGroupMembership client, group, (err, isMember) ->
        return res.status(500).send 'an error occurred'                  if err
        return res.status(400).send 'user is not a member of the group'  unless isMember
        queue.next()

    ->
      # creating a user session for the group if everything is ok
      JSession.createNewSession { username, groupName : group }, (err, session) ->
        return res.status(500).send 'an error occurred'         if err
        return res.status(500).send 'failed to create session'  unless session

        setSessionCookie res, session.clientId
        res.redirect('/')

  ]

  daisy queue


validateToken = (token, callback) ->

  Jwt.verify token, secret, { algorithms: ['HS256'] }, (err, decoded) ->
    if err
      return callback { statusCode : 400, message : 'failed to parse token' }

    unless username = decoded.username
      return callback { statusCode : 400, message : 'no username in token' }

    unless group = decoded.group
      return callback { statusCode : 400, message : 'no group slug in token' }

    return callback null, { username, group }