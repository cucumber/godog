# language: lt
@lang
@john
Savybė: lietuvis

  Raktiniai žodžiai gali būti keliomis kalbomis.

  Scenarijus: no errors event check
    Duota a feature "normal.feature" file:
    """
    Feature: the feature
      Scenario: passing scenario
        When passing step
     """
    Kai I run feature suite

    Tada the suite should have passed
    Ir the suite should have passed
    Bet the suite should have passed
