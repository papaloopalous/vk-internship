#!/usr/bin/env tarantool

local vshard = require('vshard')
local log = require('log')
local fiber = require('fiber')

box.cfg {
    listen = 3301,
    memtx_memory = 128 * 1024 * 1024,
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

vshard.router.cfg({
    bucket_count = 100,
    sharding = {
        ["shard1"] = {
            replicas = {
                ["storage1"] = {
                    uri = "storage1:3302",
                    name = "storage1",
                    master = true
                },
                ["storage2"] = {
                    uri = "storage2:3303",
                    name = "storage2"
                }
            }
        }
    }
})

log.info("Starting router configuration")

box.once("bootstrap", function()
    fiber.create(function()
        local ok, err
        -- Add initial delay to allow storages to start
        fiber.sleep(5)
        repeat
            log.info("Attempting cluster bootstrap...")
            ok, err = vshard.router.bootstrap()
            if not ok then
                log.warn("Bootstrap failed: %s. Retrying in 2 sec...", err)
                fiber.sleep(2)
            end
        until ok
        log.info("Cluster bootstrap completed successfully.")
    end)
end)

log.info("Router started at port 3301")

local function cleanup_expired_sessions()
    box.space.sessions:delete(box.space.sessions.index.expires:select({}, {iterator = 'LE', limit = 1000}))
end

fiber.create(function()
    while true do
        cleanup_expired_sessions()
        fiber.sleep(60)
    end
end)

return vshard.router
