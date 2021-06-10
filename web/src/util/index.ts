import { Currency, ActionType } from "../constants";

export const roundingPlace: (c: Currency) => number = (c) => {
  switch (c) {
    case Currency.Bitcoin:
      return 8
    case Currency.Ethereum:
      return 12
    default:
      return 8
  }
}


export const calcTotal: (
  s: IfcMarketOrder,
  a: ActionType,
  currentPrice: string,
  base: Currency,
  target: Currency,
) => number = (s, a, currentPrice, base, target) => {
  let amt = parseFloat(s.quantity)
  let price = parseFloat(currentPrice)

  switch (a) {
    case ActionType.Buy:
      if (s.base === target) {
        amt = (amt * price)
      }
      break;
    case ActionType.Sell:
      if (s.base === base) {
        amt = (amt / price)
      }
      break;
  }

  return amt
}

export const balanceMap: (b: IfcBalanceResource[]) => object = (b) => {
  const mp = {}
  for (var i = 0; i < b.length; i++) {
    mp[b[i].symbol] = parseFloat(b[i].quantity)
  }
  return mp
}