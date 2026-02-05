const generateOverride = (params = {}) => {
  let result = ''

  if (params.customCss) {
    result += `
      ${params.customCss}
    `
  }

  return result
}

module.exports = generateOverride
