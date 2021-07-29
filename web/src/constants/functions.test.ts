import { marketFromString, Currency } from './index'

it('returns a market object for a correct string', () => {
  const str = 'BTC-ETH'
  const market = marketFromString(str)

  expect(market).toStrictEqual({
    base: Currency.Bitcoin,
    target: Currency.Ethereum,
  })
})

it('returns a null value for invalid market string', () => {
  const str = 'TTT-RRR'
  const market = marketFromString(str)

  expect(market).toBe(null)
})

it('handles uppercase and lowercase properly', () => {
  const str = 'bTc-EtH'
  const market = marketFromString(str)

  expect(market).toStrictEqual({
    base: Currency.Bitcoin,
    target: Currency.Ethereum,
  })
})