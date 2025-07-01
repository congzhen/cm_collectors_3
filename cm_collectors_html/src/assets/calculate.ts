/**
 * 计算年龄的函数。
 *
 * @param _birthday 出生日期，格式为字符串（例如 "1990-01-01"）。
 * @param specifyDate 可选参数，指定的日期，用于计算相对于该日期的年龄。如果未提供或为空字符串，则使用当前日期。
 * @returns 返回根据出生日期和指定日期（或当前日期）计算出的年龄，单位为岁。
 */
export const calculateAge = function (_birthday: string, specifyDate: string | undefined = undefined) {
  // 将出生日期字符串转换为 Date 对象
  const birthday = new Date(_birthday);

  // 如果指定了有效日期，则使用该日期的时间戳；否则使用当前时间戳
  const baseDate = specifyDate && specifyDate != '' ? new Date(specifyDate).getTime() : Date.now();

  // 计算出生日期与基准日期之间的时间差（以毫秒为单位）
  const ageDifMs = baseDate - birthday.getTime();

  // 将时间差转换为 Date 对象，以便提取年份信息
  const ageDate = new Date(ageDifMs);

  // 通过 UTC 年份计算年龄，并返回绝对值
  return Math.abs(ageDate.getUTCFullYear() - 1970);
}


/**
 * 计算并返回两个数字的比率
 * 该函数用于比较宽度(w)和高度(h)的值，并以数组形式返回它们的比率
 *
 * @param w 宽度，一个数字表示
 * @param h 高度，一个数字表示
 * @returns 返回一个数组，包含宽度和高度的比率
 */
export const ratio = function (w: number, h: number) {
  // 当宽度和高度相等时，返回[1, 1]，表示比率为1:1
  if (w == h) {
    return [1, 1];
  } else if (w > h) {
    // 当宽度大于高度时，计算宽度与高度的比率，并将结果四舍五入，高度的比率为1
    return [toFixedRemore0(w / h), 1];
  } else {
    // 当高度大于宽度时，计算高度与宽度的比率，并将结果四舍五入，宽度的比率为1
    return [1, toFixedRemore0(h / w)];
  }
}
/**
 * 将数字转换为指定小数位数的字符串，并移除末尾的零
 * 这个函数的目的是格式化数字输出，使其在视觉上更加简洁
 * 例如，将数字1.5转换为小数点后两位并移除末尾的零，结果应为"1.5"而不是"1.50"
 *
 * @param num 要格式化的数字
 * @param fixed 小数点后的位数，默认为2
 * @returns 格式化后的字符串，不包含末尾的零
 */
const toFixedRemore0 = function (num: number, fixed = 2) {
  // 将数字转换为指定小数位数的字符串
  // 然后移除末尾的零，如果小数点后全是零，则移除小数点
  return parseFloat(num.toFixed(fixed).toString().replace(/\.?0+$/, ''));
}
