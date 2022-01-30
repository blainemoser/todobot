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
