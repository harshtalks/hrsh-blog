export namespace Result {
  export type Ok<T> = {
    readonly _tag: 'Ok'
    readonly value: T
  }

  export type Err<E> = {
    readonly _tag: 'Err'
    readonly error: E
  }

  export type Result<T, E> = Ok<T> | Err<E>

  export const Ok = <T>(value: T): Ok<T> => ({
    _tag: 'Ok',
    value
  })

  export const Err = <E>(error: E): Err<E> => ({
    _tag: 'Err',
    error
  })

  export const isOk = <T, E>(result: Result<T, E>): result is Ok<T> => {
    return result._tag === 'Ok'
  }

  export const isErr = <T, E>(result: Result<T, E>): result is Err<E> => {
    return result._tag === 'Err'
  }

  export const match = <T, E, OkResult, ErrResult>(
    result: Result<T, E>,
    onOk: (value: T) => OkResult,
    onErr: (error: E) => ErrResult
  ): OkResult | ErrResult => {
    if (isOk(result)) {
      return onOk(result.value)
    }

    return onErr(result.error)
  }

  export type MaybeResult<T,E> = T extends Promise<infer U> ?  Promise<Result<U,E>> : Result<T,E>
}

type Callback = (...value: any) => any

export const useAsync = <TCb extends Callback>(cb: TCb) => {
  return <Err extends any = unknown,Value extends ReturnType<TCb> = ReturnType<TCb>> (...args: Parameters<TCb>): Result.MaybeResult<Value,Err> => {
    try {
      const response = cb(...args)
      if (response instanceof Promise) {
        return response.then(Result.Ok).catch(Result.Err) as unknown as Result.MaybeResult<Value,Err>
      }

      return Result.Ok(response) as  Result.MaybeResult<Value,Err>
    } catch (e) {
     return Result.Err(e) as  Result.MaybeResult<Value,Err>
    }
  }
}

const promiseString = (value: string) => Promise.resolve(value)

const asyncString = await useAsync(promiseString)(`Hello, World!`)


const x = Result.match(asyncString)(v =>v, e => e)
