
# package 说明

## memo_test
`src/learn/memo`

实现结果缓存的功能，5种方案逐步演进：
- 不考虑多线程的实现，有竞态问题
- 使用互斥量，但是退化为顺序执行
- 使用互斥量，只将memo的访问放入临界区
- 使用互斥量+channel，解决方案二和三的问题，但是引入了重复执行缓存操作的问题
- 使用无锁的channel方案

## format
`src/learn/format`
- 输出数据详细内部结构，暂不支持循环引用
- 转为Lisp语言的S表达式，类似于JSON、XML，它也是一种格式

## deepequal
在`reflect.DeepEqual`的功能上扩展，`nil slice`和值不为`nil`的空`slice`也相等，`map`类似。