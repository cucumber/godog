@dupSteps
Feature:	Dupsteps example features

  @flatTire
  Scenario: Flat Tire
    Given I ran over a nail and got a flat tire
    Then I fixed it
    Then I can continue on my way

  @cloggedDrain
  Scenario: Clogged Drain
    Given I accidentally poured concrete down my drain and clogged the sewer line
    Then I fixed it
    Then I can once again use my sink
