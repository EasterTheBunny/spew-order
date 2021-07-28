declare module 'web-worker:*' {
  const WorkerFactory: new () => Worker;
  export default WorkerFactory;
}

interface IfcWorkerMessage {
  type: string
}

interface IfcTickerMessage extends IfcWorkerMessage {
  price: string
  high_24h: string
  low_24h: string
  open_24h: string
}

interface IfcBookMessage extends IfcWorkerMessage {
  maxDepth: number
  asks: string[][]
  bids: string[][]
}