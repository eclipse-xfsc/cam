# Authentication Security Collection Module

A collection service to gather information on an OAuth 2.0/2.1 / OpenID Connect AuthN/AuthZ service.
Gathers the following information:

- Supported Authorization Grants and Server Metadata in general
- Implemented SecBCP Mitigations for the above
- Feasibility of basic Authorization Flows for the above grants (As long as no user interaction is required)
- Implementation of a range of optional features

## Necessary Information for Operation

- Issuer identifier
- Discovery: (Pick one)
    - Automated Discovery (Boolean)
        - If the Identifier contains a path, an indication whether legacy OpenID Connect Discovery should be used to locate the Metadata Document. (Bool)
    - Server Metadata Document (URL, Can be inferred from the above, if implemented according to standards)
    - Explicit Configuration: (Not recommended, may expand, Implies no checks for Discovery)
        - List of Grants to check (Array of Strings)
        - Token Endpoint (URL, if any grants need it)
        - Authorization Endpoint (URL, if any grants need it)
        - JWKS Location (URL, if the server supports any signing algorithms)
        - PAR Endpoint (URL, if supported)
        - Userinfo Endpoint (URL, if supported)
- Client Registration: (Pick one)
    - Dynamic Registration (Boolean)
    - Explicit Configuration (Possibly one for each Auth method, Must be registered at the service):
        - Client ID (String)
        - Authentication Method (String)
            - If `client_secret_{post,basic,jwt}`: The Secret (String)
            - If `private_key_jwt`: The private key (Key)

