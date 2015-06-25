# language: lt
@lang
Savybė: užkrauti savybes
  Kad būtų galima paleisti savybių testus
  Kaip testavimo įrankis
  Aš turiu galėti užregistruoti savybes

  Scenarijus: savybių užkrovimas iš aplanko
    Duota savybių aplankas "features"
    Kai aš išskaitau savybes
    Tada aš turėčiau turėti 4 savybių failus:
      """
      features/events.feature
      features/lang.feature
      features/load.feature
      features/run.feature
      """
