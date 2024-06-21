Feature: Intercepting steps in order to manipulate the response
	This test suite installs an interceptor that perfoms some manipulation on the test step result.
	The manipulation in this test is arbitrary and for illustration purposes.
	The interceptor inverts passing/failing results of any steps containing the text FLIP_ME.

  Scenario: The trigger word is not present so a pass status should be untouched
    When passing step should be passed

  Scenario: The trigger word is not present so a fail status should be untouched
    When failing step should be failed

  Scenario: Trigger word should should flip a fail to a pass
    When failing step with the word FLIP_ME should be passed

  Scenario: Trigger word should should flip a pass to a fail
    When passing step with the word FLIP_ME should be failed
