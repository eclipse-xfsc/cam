Feature: Configuration Interface

    In the configuration interface, users select the services they want to monitor,
    and the controls they want to apply, for instance a user may select a storage
    service that she uses, and select a control regarding at-rest encryption to be
    monitored on that service.

    @CI-F-01
    @Backend
    Scenario: Services and Controls

        Given  the following services exist:
            | ID                                   | name       | description    |
            | C5FE4694-E2D4-4028-94B9-BDBADF2921AF | My Service | My Description |
        When the user accesses "/v1/configuration/cloud_services"
        Then it should present a list of monitorable services

    @CI-F-02
    @Backend
    Scenario: Configure Collection Modules

        The configuration user interface must offer the user the possibility to configure
        the collection modules where applicable, e.g., when a collection module requires
        a credential for accessing an API

        Given  the following services exist:
            | ID                                   | name       | description    |
            | C5FE4694-E2D4-4028-94B9-BDBADF2921AF | My Service | My Description |
        When the user accesses "/v1/configuration/cloud_services/{service_id}/configurations" with "service_id" set to "C5FE4694-E2D4-4028-94B9-BDBADF2921AF"
        Then he can configure the cloud service "C5FE4694-E2D4-4028-94B9-BDBADF2921AF"
