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
