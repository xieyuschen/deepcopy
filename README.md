# DeepCopy

![example workflow](https://github.com/xieyuschen/deepcopy/actions/workflows/go.yml/badge.svg)
[![codecov](https://codecov.io/gh/xieyuschen/deepcopy/branch/master/graph/badge.svg?token=E1IU1FAK92)](https://codecov.io/gh/xieyuschen/deepcopy)

The repo is original at [mohae/deepcopy](github.com/mohae/deepcopy). As it hasn't been maintained for a long time, it's
maintained here with bug fix and new features.

## Expected Behaviors in DeepCopy Library

This topic lists some expected behaviors in `DeepCopy` library here.

| Type             | Expected Behavior                                                                                                                                                                                                                                                                                                                                            |
| :--------------- | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Primitive Types  | Copy by the value                                                                                                                                                                                                                                                                                                                                            |
| Struct           | Construct a new struct, deep copy each field according to the respective behaviors                                                                                                                                                                                                                                                                           |
| Slice/Array      | Construct a new slice/array, deep copy underlying types according to the respective behaviors                                                                                                                                                                                                                                                                |
| Interface        | Construct a new interface, deep copy underlying types according to the respective behaviors                                                                                                                                                                                                                                                                  |
| Map              | Construct a new map, deep copy underlying types according to the respective behaviors                                                                                                                                                                                                                                                                        |
| Channel          | Channel will be shared between source and copied.                                                                                                                                                                                                                                                                                                            |
| Function         | Function will be shared betwwen source and copied. Beware the variables captured by your function when you do deep copy                                                                                                                                                                                                                                      |
| Pointer          | Construct a new object with the type under the pointer according to the respective behaviors and then return the a pointer to the new object                                                                                                                                                                                                                 |
| time.Time        | A new time.Time object will be constructed with the shared `*time.Location` pointer inside it                                                                                                                                                                                                                                                                |
| Unexposed Fields | Unexposed fields are not supported in reflect, as a result, it cannot copy the inside status. <br/>Accroding to this, states of objects cannot be conserved and some of objects cannot be used at all. <br/>For example, a mutex will be deep copied like a value because all internal states are lost during deep copy. A `*os.File` cannot be used at all. |

## Acknowledge

There was a [deep copy proposal](https://github.com/golang/go/issues/51520) in Go but got declined due to no consensus.
The brief of arguments are listed below.

- The deep copy should be a library, instead of go language itself.  
  "ianlancetaylor:_... the best way forward here is going to be to write the version of the function that you think we need. And that implementation can live anywhere..."_

- The behaviors of deep copy for circular/recursive data structure.
- The behaviors of deep copy for stateful structures such as `*os.File`, `mutex`, `channel` and so on.
- ...

As the deep copy function shouldn't belong to go language, and it's useful in some cases, the library came into being.

## Contribute

This repo will be maintained in a long term even though the original author start to maintain the original repo. I will
try to fix any unexpected bug and add some features if necessary. Glad to receive issues and PRs.

Thank [Joel Scoble](https://github.com/mohae), [Mathieu Champlon](https://github.com/mat007), [Damien Neil](https://github.com/neild), [Sergey Cherepanov](https://github.com/cheggaaa) for their previous work.
