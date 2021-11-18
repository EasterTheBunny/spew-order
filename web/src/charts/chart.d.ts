
interface PriceDepthItem {
  price: string
  depth: number
}

interface CandleItem {
  time: Date
  low: number
  high: number
  open: number
  close: number
  volume: number
}

interface AssetItem {
  name: string
  nominal: number
  amount: number
}

interface PriceDepthChart {
  draw: (width: number, height: number, color: string, yAxis: boolean) => void
  update: (items: PriceDepthItem[], depth: number) => void
  remove: () => void
}

interface PriceHistoryChart {
  draw: (width: number, height: number) => void
  update: (items: CandleItem[]) => void
  remove: () => void
}

interface AssetChart {
  draw: (height: number) => void
  update: (items: AssetItem[]) => void
  remove: () => void
}