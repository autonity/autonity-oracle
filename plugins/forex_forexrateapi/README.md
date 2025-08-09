# ForexAPI Oracle Plugin

## Overview

This plugin integrates with **ForexRateAPI.com** to fetch foreign exchange (forex) rate data for the Autonity Oracle Server. It acts as a data adaptor, allowing the oracle to source reliable and real-time currency exchange rates from a professional data provider.

The plugin is designed to be lightweight and efficient, fetching the latest rates for a given set of currency pairs.

## Data Source

- **Provider**: ForexRateAPI.com
- **Official Site**: <https://forexrateapi.com/>
- **Data Quality**: The API provides real-time and historical data for 200+ currencies, updated frequently. It aggregates data from multiple reputable sources to ensure accuracy and reliability.

---

## Service Key Subscription

An API key is **required** to use this plugin. Follow these steps to obtain one:

1. **Sign Up**: Go to the [ForexRateAPI.com](https://forexrateapi.com/) website and sign up for an account.
2. **Choose a Plan**: ForexRateAPI offers several subscription plans. For reliable oracle operation, we recommend at least the **Basic Plus** plan. This plan provides a sufficient number of monthly API requests to support frequent price updates required by a production oracle server. The free plan has a very low request limit and is only suitable for initial testing.
3. **Get API Key**: Once you have subscribed to a plan, you will find your unique API key in your user dashboard.


## Plugin Configuration

To use this plugin, add its configuration under the `pluginConfigs:` section in your oracle server's main configuration file (`oracle_config.yml`).

Here is a commented example configuration block to guide you:

```yaml
# In your main oracle server configuration file, under pluginConfigs:

  - name: forex_forexrateapi                  # required, it is the plugin file name in the plugin directory.
    key: 6ec1e92.....123abc                   # required, visit [https://forexrateapi.com](https://forexrateapi.com) to get your key, IMPORTANT: do not use free or developer service plan.
    refresh: 300                              # optional, buffered data within 300s, recommended for API rate limited data source.
````

-----

## Build and Use

1.  **Build the Plugin**: Navigate to the root `autonity-oracle` directory and run the build command:
    ```shell
    go build -o ./build/bin/plugins/forex_forexrateapi ./plugins/forex_forexrateapi/forex_forexrateapi.go
    ```
    This will create a binary file named `forex_forexrateapi` inside the `./build/bin/plugins/` directory.
2.  **Run**: The oracle server will automatically discover and load the plugin from the `plugins` directory upon startup. Ensure your configuration file is correctly set up as described above.

```