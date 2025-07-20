#!/usr/bin/env tarantool

local vshard = require('vshard')
local log = require('log')

box.cfg {
    listen = 3302,
    memtx_memory = 128 * 1024 * 1024,
    vinyl_memory = 128 * 1024 * 1024,
    replication = {'storage2:3303'},
    work_dir = '/var/lib/tarantool',
    replication_connect_timeout = 30
}

box.once('sessions_space_init', function()
        local sessions = box.schema.space.create('sessions', {
            format = {
                {name = 'session_id', type = 'string'},
                {name = 'user_id',   type = 'string'},
                {name = 'role',      type = 'string'},
                {name = 'expires_at',type = 'number'}
            },
            if_not_exists = true
        })

        sessions:create_index('primary', {
            parts = {{field = 'session_id', type = 'string'}},
            if_not_exists = true
        })

        sessions:create_index('expires', {
            parts = {{field = 'expires_at', type = 'number'}},
            if_not_exists = true
        })

        box.schema.user.grant('guest', 'read,write',
                               'space', 'sessions', nil,
                               {if_not_exists = true})
    end)

vshard.storage.cfg({
    bucket_count = 100,
    sharding = {
        ['shard1'] = {
            replicas = {
                ['storage1'] = {
                    uri = 'storage1:3302',
                    name = 'storage1',
                    master = true
                },
                ['storage2'] = {
                    uri = 'storage2:3303',
                    name = 'storage2'
                }
            }
        }
    }
}, 'storage1')

log.info("Storage1 started at port 3302")
