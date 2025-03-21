from alpaca.trading.requests import GetAssetsRequest
from alpaca.trading.enums import AssetClass

from src.alpaca.client import trading_client

account = trading_client.get_account()

def fetch_assets(asset_class=AssetClass.CRYPTO):
    search_params = GetAssetsRequest(asset_class=asset_class)
    assets = trading_client.get_all_assets(search_params)
    return assets
