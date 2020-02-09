from setuptools import setup, find_packages


setup(
    name="maker_simulator",
    packages=find_packages(),
    install_requires=[
        # "web3", use forked version
    ],
    scripts=["./bin/maker-simulator"],
    extras_require={
        "dev": [
            "pylint",
        ],
    },
)
