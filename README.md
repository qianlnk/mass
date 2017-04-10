# MASS

[![Build Status](https://travis-ci.org/qianlnk/mass.svg?branch=master)](https://travis-ci.org/qianlnk/mass)

用用户提供的`processing method`生产ID，生产ID时确保在本次生产的ID完成前同一组参数不再生产其它不一致的ID（解决并发、重试、刷机、多机部署等问题），另外查库插库需要用户在`processing method`中自己实现，以确保所生产的ID参数在数据库中不重复。

## The main technical stack

* redis实现分布式锁
* redis订阅发布
* 生产频率限制

## API

`processing method`

```golang
type ProcessingMethod func(materials ...interface{}) interface{}
```

## Usage

see demo