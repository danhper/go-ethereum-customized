from setuptools import setup, find_packages


setup(
    name="feed_syncer",
    packages=find_packages(),
    install_requires=[
        # "web3", use forked version
    ],
    extras_require={
        "dev": [
            "pylint",
        ],
    },
)
