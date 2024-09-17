# 配置模型映射

## 模型介绍

模型映射决定了用户请求模型和官网实际模型的对应关系

模型映射使用`json`格式配置,以下为推荐示例:

```json
{
  "auto": "auto",
  "gpt-3.5-turbo": "gpt-4o-mini",
  "gpt-4o-lite": "gpt-4o-lite",
  "gpt-4o-mini": "gpt-4o-mini",
  "gpt-4": "gpt-4",
  "gpt-4o": "gpt-4o",
  "o1-preview": "o1-preview",
  "o1-mini": "o1-mini"
}
```
::: tip
在模型映射中,`key`为用户请求的模型名称,`value`为官网实际模型名称
:::

即止目前为止,官网支持的模型有:

免费账号: `auto`, `gpt-4o-mini`,  `gpt-4o`(在本系统中转换为`gpt-4o-lite`)

付费账号: `gpt-4`, `gpt-4o`, `o1-preview`, `o1-mini`, `gpt-4o-mini`(付费账号不会调用这个模型,如需要调用请在后台添加免费账号)

::: tip
因为免费账号和付费账号都有模型gpt-4o,所以在模型映射中,免费账号映射为`gpt-4o-lite`,付费账号映射为`gpt-4o`以使系统可以正确识别调用对应的账号
:::

## 配置方法 

进入管理后台 `http://服务器地址:8100/xyhelper` 登陆

在 `系统管理` -> `参数配置` -> `参数列表` 中 新建 `modelmap` 参数,并将上述示例内容复制到 `参数值` 中,点击保存即可

