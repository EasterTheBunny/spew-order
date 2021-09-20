import type { AxiosResponse } from "axios"
import axios, { AxiosInstance, AxiosRequestConfig } from "axios"

export default class ExchangeAPI {
  private api: AxiosInstance
  private options: AxiosRequestConfig = {
    timeout: 5000,
    headers: {
      "Content-Type": "application/vnd.api+json",
    }
  }

  private static ACCOUNT_PATH: string = "/accounts"
  private static ORDER_PATH: string = "/orders"
  private static TRANSACTION_PATH: string = "/transactions"
  private static ADDRESS_PATH: string = "/addresses"

  constructor(url: string) {
    this.options = Object.assign({}, this.options, { baseURL: url })
    this.api = axios.create(this.options)
  }

  public getActiveAccountFunc(): (accountID: string) => Promise<IfcAccountResource> {
    const f: (inst: AxiosInstance) => (accountID: string) => Promise<IfcAccountResource> = (inst) => {
      return async (accountID) => {
        if (accountID !== null && accountID !== "") {
          return inst.get(ExchangeAPI.ACCOUNT_PATH+"/"+accountID).then((x) => this.dataResponse<IfcAccountResource>(x))
        } else {
          return inst.get(ExchangeAPI.ACCOUNT_PATH)
                        .then((x) => {
                          let y = this.dataResponse<IfcAccountResource[]>(x)
                          return inst.get(ExchangeAPI.ACCOUNT_PATH+"/"+y[0].id).then((r) => {
                            return this.dataResponse<IfcAccountResource>(r)
                          })
                        })
        }
      }
    }

    return f(this.api)
  }

  public getOrderFunc(): (accountID: string, data: IfcOrderResource) => Promise<IfcOrderResource[] | IfcOrderResource> {
    const f: (inst: AxiosInstance) => (accountID: string, data: IfcOrderResource) => Promise<IfcOrderResource[] |IfcOrderResource> = (inst) => {
      return async (accountID, data) => {
        const path = ExchangeAPI.ACCOUNT_PATH+"/"+accountID+ExchangeAPI.ORDER_PATH

        if (data !== null && data.guid === "") {
          // post new
          return inst.post(path, data.order).then((x) => this.dataResponse<IfcOrderResource[]>(x))
        } else if (data !== null && data.guid !== "") {
          if (data.status != "") {
            // patch status
            // hard coded patch order as this is the only support type
            return inst.patch(path+"/"+data.guid, [{ op: "replace", path: "/status", value: data.status}]).then((x) => this.dataResponse<IfcOrderResource[]>(x))
          } else {
            // get by id
            return inst.get(path+"/"+data.guid).then((x) => this.dataResponse<IfcOrderResource[]>(x))
          }
        } else {
          // get all values
          return inst.get(path).then((x) => this.dataResponse<IfcOrderResource[]>(x))
        }
      }
    }

    return f(this.api)
  }

  public getTransactionFunc(): (accountID: string, data: IfcTransactionRequest) => Promise<IfcTransactionResource[] | IfcTransactionResource> {
    const f: (inst: AxiosInstance) => (accountID: string, data: IfcTransactionRequest) => Promise<IfcTransactionResource[] | IfcTransactionResource> = (inst) => {
      return async (accountID, data) => {
        const path = ExchangeAPI.ACCOUNT_PATH+"/"+accountID+ExchangeAPI.TRANSACTION_PATH

        if (data !== null) {
          // post new
          return inst.post(path, data).then((x) => this.dataResponse<IfcTransactionResource>(x))
        } else {
          // get all values
          return inst.get(path).then((x) => this.dataResponse<IfcTransactionResource[]>(x))
        }
      }
    }

    return f(this.api)
  }

  public getAddressFunc(): (accountID: string, symbol: string) => Promise<IfcAddressResource> {
    const f: (inst: AxiosInstance) => (accountID: string, symbol: string) => Promise<IfcAddressResource> = (inst) => {
      return async (accountID, symbol) => {
        const path = ExchangeAPI.ACCOUNT_PATH+"/"+accountID+ExchangeAPI.ADDRESS_PATH+"/"+symbol
        return inst.get(path).then((x) => this.dataResponse<IfcAddressResource>(x))
      }
    }

    return f(this.api)
  }

  public async setBearerToken(token: string) {
    this.api.defaults.headers.common['Authorization'] = "Bearer " + token
  }

  private dataResponse<T>(r: AxiosResponse<IfcAPIResponse<T>>): T {
    return this.extractData<T>(r.data)
  }

  private extractData<T>(r: IfcAPIResponse<T>): T {
    const d: T = r.data
    return d
  }
}