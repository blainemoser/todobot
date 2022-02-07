package tests

const TestEventPayload = `{
    "token": "vcSe2kbpQsFkpBJyVdnM4o5M",
    "team_id": "T02CCAKL1JB",
    "api_app_id": "A02NLFZUPB7",
    "event": {
        "client_msg_id": "202c5873-3f9c-45d9-b64b-4b6c90ba746c",
        "type": "app_mention",
        "text": "<@U02P23SK0SV> [message]",
        "user": "U02DGLZ7ABA",
        "ts": "[timestamp]",
        "team": "T02CCAKL1JB",
        "blocks": [
            {
                "type": "rich_text",
                "block_id": "oJh2",
                "elements": [
                    {
                        "type": "rich_text_section",
                        "elements": [
                            {
                                "type": "user",
                                "user_id": "U02P23SK0SV"
                            },
                            {
                                "type": "text",
                                "text": " Her there."
                            }
                        ]
                    }
                ]
            }
        ],
        "channel": "C02NLG80TEH",
        "event_ts": "1643038674.000200"
    },
    "type": "event_callback",
    "event_id": "Ev02VBF8EFPD",
    "event_time": 1643038674,
    "authorizations": [
        {
            "enterprise_id": null,
            "team_id": "T02CCAKL1JB",
            "user_id": "U02P23SK0SV",
            "is_bot": true,
            "is_enterprise_install": false
        }
    ],
    "is_ext_shared_channel": false,
    "event_context": "4-eyJldCI6ImFwcF9tZW50aW9uIiwidGlkIjoiVDAyQ0NBS0wxSkIiLCJhaWQiOiJBMDJOTEZaVVBCNyIsImNpZCI6IkMwMk5MRzgwVEVIIn0"
}`

const TestUserPayload = `{
    "ok": true,
    "user": {
        "id": "%s",
        "team_id": "T02CCAKL1JB",
        "name": "blainemoser",
        "deleted": false,
        "color": "9f69e7",
        "real_name": "blainemoser",
        "tz": "Africa/Harare",
        "tz_label": "Central Africa Time",
        "tz_offset": 7200,
        "profile": {
            "title": "",
            "phone": "",
            "skype": "",
            "real_name": "blainemoser",
            "real_name_normalized": "blainemoser",
            "display_name": "",
            "display_name_normalized": "",
            "fields": null,
            "status_text": "",
            "status_emoji": "",
            "status_emoji_display_info": [],
            "status_expiration": 0,
            "avatar_hash": "g7709d97b936",
            "first_name": "blainemoser",
            "last_name": "",
            "image_24": "",
            "image_32": "",
            "image_48": "",
            "image_72": "",
            "image_192": "",
            "image_512": "",
            "status_text_canonical": "",
            "team": "T02CCAKL1JB"
        },
        "is_admin": true,
        "is_owner": true,
        "is_primary_owner": true,
        "is_restricted": false,
        "is_ultra_restricted": false,
        "is_bot": false,
        "is_app_user": false,
        "updated": 1630334725,
        "is_email_confirmed": true,
        "who_can_share_contact_card": "EVERYONE"
    }
}`
