Feature: Attaching content to the cucumber report
  The cucumber JSON and NDJSON support the inclusion of attachments.
  These can be text or images or any data really. 

  Scenario: Attaching files to the report
    Given I have attached two documents in sequence 
    And I have attached two documents at once
