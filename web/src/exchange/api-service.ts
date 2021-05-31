import axios, { AxiosInstance, AxiosRequestConfig } from "axios"
import type { AxiosResponse } from "axios"
import type { User } from "oidc-client"

export default class ExchangeAPI {
  private api: AxiosInstance
  private options: AxiosRequestConfig = {
    timeout: 5000,
    headers: {
      "Content-Type": "application/vnd.api+json",
    }
  }

  private static ACCOUNT_PATH: string = "/account"

  constructor(url: string) {
    this.options = Object.assign({}, this.options, { baseURL: url })
    this.api = axios.create(this.options)
  }

  public getActiveAccountFunc(): (str: string) => Promise<IfcAccountResource> {
    const f: (inst: AxiosInstance) => (str: string) => Promise<IfcAccountResource> = (inst) => {
      return async (str) => {
        return inst.get(ExchangeAPI.ACCOUNT_PATH)
                      .then((x) => {
                        let y = this.dataResponse<IfcAccountResource[]>(x)
                        return inst.get(ExchangeAPI.ACCOUNT_PATH+"/"+y[0].id).then((r) => {
                          return this.dataResponse<IfcAccountResource>(r)
                        })
                      })
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