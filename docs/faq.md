# 用户常见疑问

## ERR_CONNECTION_CLOSED / 代理关闭且没有日志

常见原因：

- 密码错误，请检查您的密码是否正确。

## 好慢

网络速度与线路质量有关，升级更优质的线路可以提升速度。

## 能不能不要 TLS

不能。

## 能不能过 CDN

不支持 CDN，因为添加 ws 等传输层会影响代理的混淆效果。此外，某些 CDN 服务已对这类滥用进行了限制。

## 这和 xtls-vision 有什么区别

本项目没有写死的填充策略，没有代理数据 TLS v1.3 握手完毕后直接拷贝到 TCP。有降低代理握手延迟的连接复用。

## 这和 h2mux 有什么区别

本项目有灵活可调的包长分布策略，更低的性能开销，更好的连接复用策略。

## 为什么选项这么少

本项目只是提供一个简洁的 Any in TLS 代理的示例，并不旨在成为“通用代理工具”。

## 为什么是自签证书

近几年没见过 GFW 进行 MITM 攻击或者因为主动探测到自签证书而墙 IP。

作为参考实现，不对 TLS 协议本身做过多的处理。

## FingerPrint 之类的选项呢

TLS 本身（ClientHello）的特征不是本项目关注的重点，改变这些特征要容易的多。

- 某些 Golang 灵车代理早已标配 uTLS，为什么还被墙？
- 某些 Golang 灵车代理即使不用 uTLS，直接使用 Golang TLS 栈，为什么不被墙？

## 关于默认的 PaddingScheme

默认 PaddingScheme 只是一个示例。本项目无法确保默认参数不会被墙，因此设计了通过更新参数改变流量特征的机制。

## 如何更改 PaddingScheme

服务器设置 `--padding-scheme ./padding.txt` 参数。

## 还有别的 PaddingScheme 吗

模拟 XTLS-Vision:

```
stop=3
0=900-1400
1=900-1400
2=900-1400
```

模仿的不是特别像，但可以说明 XTLS-Vision 的弊端：写死的长度处理逻辑，只要 GFW 更新特征库就能识别。

## 为什么只处理上行

- 回国带宽珍贵
- 前期测试发现处理下行基本没有效果
- 如果你想处理下行，自己改服务器代码就可以了

## 参考过的项目

https://github.com/xtaci/smux

https://github.com/3andne/restls

https://github.com/SagerNet/sing-box
