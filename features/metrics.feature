Feature: Metrics and Controls

    @MC-C-01
    @Model
    Scenario: Load OSCAL Model

        Controls must be persisted using the properties specified by the OSCAL model

        Given an OSCAL model on file
        When it is loaded
        Then it should contain requirements

    @MC-C-02
    @Model
    Scenario: Metric Format

        Metrics must be persisted in the format specified in this document (see below) and be
        transmitted in the format specified in 3.3.1.

        Given an OSCAL model on file
        When it is loaded
        Then the control "OPS-21" should have metric "SystemComponentsIntegrity"
        Then the metric "SystemComponentsIntegrity" should have scale 1
