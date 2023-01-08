# Knweb

# Features

已完成
+ Server抽象
+ 路由：简单路由匹配、通配符匹配、带参数的路由匹配、正则匹配
+ AOP
+ 简单Context上下文管理，包括请求的输入、输出
  + 输入： 
    + Body输入, 例如提供ctx.Req.Body.BindJson()
    + 表单输入，例如，提供FormValue()
    + 查询参数/路径参数/StringValue，例如提供QueryValue()/PathValue()的实现。为了防止每次请求都解析Query，在Query增加了缓存
  + 输出： 
    + 返回Resp进行反序列化。例如，提供RespJSON的实现
    + 通过go template对响应进行render

TODO
> + 可路由的AOP