import beneath

from generators import earthquakes

with open("schemas/earthquake.graphql", "r") as file:
    EARTHQUAKES_SCHEMA = file.read()

if __name__ == "__main__":
    p = beneath.Pipeline(parse_args=True)
    earthquakes = p.generate(earthquakes.generate_earthquakes)
    p.write_stream(
        earthquakes,
        "earthquakes",
        schema=EARTHQUAKES_SCHEMA,
        description="Earthquakes fetched from https://earthquake.usgs.gov/",
    )
    p.main()
