/**
 * 防抖函数
 *
 * @param func 要执行的函数
 * @param duration 延迟时间，单位毫秒
 */
export const debounce = (func: Function, duration = 500) => {
    let timerId: number | undefined;
    return (...args: any[]) => {
        clearTimeout(timerId);
        timerId = Number(setTimeout(() => {
            func.apply(this, args);
        }, duration));
    };
};

/**
 * 防抖函数（立即执行版）
 *
 * @param func 要执行的函数
 * @param wait 延迟时间，单位毫秒
 */
export const debounceNow = (func: Function, wait = 500) => {
    let timerId: number | undefined;
    return (...args: any[]) => {
        clearTimeout(timerId);
        const callNow = !timerId;
        timerId = Number(setTimeout(() => {
            timerId = undefined;
        }, wait));
        if (callNow) {
            func.apply(this, args);
        }
    };
};

/*
*时间戳版和定时器版的节流函数的区别就是，时间戳版的函数触发是在时间段内开始的时候，而定时器版的函数触发是在时间段内结束的时候。
*/

/**
 * 节流函数（时间戳版）
 * 指连续触发事件但是在 n 秒中只执行一次函数。即 2n 秒内执行 2 次... 。节流如字面意思，会稀释函数的执行频率。
 * 
 * @param func 要执行的函数
 * @param wait 延迟时间，单位毫秒
 */
export const throttle = (func: Function, wait = 1) => {
    let previous = 0;
    return (...args: any[]) => {
        const now = Date.now();
        if (now - previous > wait) {
            func.apply(this, args);
            previous = now;
        }
    };
};

/**
 * 节流函数（定时器版）
 * 在持续触发事件的过程中，函数不会立即执行，并且每 1s 执行一次，在停止触发事件后，函数还会再执行一次。
 * 
 * @param func 要执行的函数
 * @param wait 延迟时间，单位毫秒
 */
export const throttleTimeout = (func: Function, wait = 1) => {
    let timeoutId: number | undefined;
    return (...args: any[]) => {
        if (!timeoutId) {
            timeoutId = Number(setTimeout(() => {
                timeoutId = undefined;
                func.apply(this, args);
            }, wait));
        }
    };
};