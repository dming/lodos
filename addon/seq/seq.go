//seq means sequence..imitate [wechat] structure

package seq

import "github.com/dming/lodos/addon/seq/seqconf"

//[16]byte as uuid, plus int64 format [16]byte as sequence, as [32]byte
type Sequence int64 //can be formatted as [16]byte
type UserSequence [32]byte //[32]byte, as format [uuid][int64]

//uid is int, it should be check from db->map[uuid]int
type Section interface {
	GetFlag() int
	AddUser(uid int, seq Sequence) error
	Exist(uid int) bool
	GetSeq(uid int) (seq Sequence, err error)
	//UpdateMaxSeq()
	GetMaxSeq() Sequence
}

type SectionManager interface {
	//begin id
	//size
	//section list
}

//Run in Alloc
/**
AllocSvr
LeaseCherk
SectionManager
Section
 */
//include a store_svr for load route from db, and save to db
//a section means a sequence service unit,so it need a section_manager
//a section_manager contains continuous id sections
type AllocSvr interface {
	LeaseClerk

	OnInit(settings map[string]string) //assign settings and init from newest route table
	//Get request from client and return seq [or route table in addition]
	//client request should include user ID and route version
	Run()
	OnDestroy()

	GetSeq(id int) (seq Sequence, err error)
}

/*
   所谓租约(leases)，其实就是一个合同，即服务端给予客户端在一定期限内可以控制修改操作的权力。
   如果服务端要修改数据，首先要征求拥有这块数据的租约的客户端的同意，之后才可以修改。
   客户端从服务端读取数据时往往就同时获取租约，在租约期限内，如果没有收到服务端的修改请求，
   就可以保证当前缓存中的内容就是最新的。如果在租约期限内收到了修改数据的请求并且同意了，
   就需要清空缓存。在租约过期以后，客户端如果还要从缓存读取数据，就必须重新获取租约，
   我们称这个操作为“续约”{!引用自<<租约机制简介>>}。
 */

// ## 租约
// 为了避免失联AllocSvr提供错误的服务，返回脏数据，AllocSvr需要跟StoreSvr保持租约。
// 这个租约机制由以下两个条件组成：
// > - 租约失效：AllocSvr N秒内无法从StoreSvr读取加载配置时，AllocSvr停止服务
//   - 租约生效：AllocSvr读取到新的加载配置后，立即卸载需要卸载的号段，
//     需要加载的新号段等待N秒后提供服务
//
// 这两个条件保证了切换时，新AllocSvr肯定在旧AllocSvr下线后才开始提供服务。
// 但这种租约机制也会造成切换的号段存在小段时间的不可服务，
// 不过由于微信后台逻辑层存在重试机制及异步重试队列，
// 小段时间的不可服务是用户无感知的，而且出现租约失效、切换是小概率事件，整体上是可以接受的。
//
type LeaseClerk interface {
	// 租约生效
	OnLeaseValid() //callback
	// 路由表更新
	OnLeaseUpdated() //callback
	// 租约失效
	OnLeaseInvalid() //callback

	// 开始租约服务
	Start()
	// 停止租约服务
	Stop()
}

// ## 容灾2.0架构：嵌入式路由表容灾
// - 既然Client端与AllocSvr存在路由状态不一致的问题，
//   那么让AllocSvr把当前的路由状态传递给Client端，
//   打破之前只能根据本地Client配置文件做路由决策的限制，从根本上解决这个问题。
//
// - 所以在2.0架构中，我们把AllocSvr的路由状态嵌入到Client请求sequence的响应包中，
//   在不带来额外的资源消耗的情况下，实现了Client端与AllocSvr之间的路由状态一致。
//
// ## 具体实现方案如下：
//
// seqsvr所有模块使用了统一的路由表，描述了uid号段到AllocSvr的全映射。
// 这份路由表由仲裁服务根据AllocSvr的服务状态生成，写到StoreSvr中，
// 由AllocSvr当作租约读出，最后在业务返回包里旁路给Client端。
//
// 路由表优化：
// 1. Client根据本地共享内存缓存的路由表，选择对应的AllocSvr；
//    如果路由表不存在，随机选择一台AllocSvr
//
// 2. 对选中的AllocSvr发起请求，请求带上本地路由表的版本号
//
// 3. AllocSvr收到请求，除了处理sequence逻辑外，
//    判断Client带上版本号是否最新，如果是旧版则在响应包中附上最新的路由表
//
// 4. Client收到响应包，除了处理sequence逻辑外，判断响应包是否带有新路由表。
//    如果有，更新本地路由表，并决策是否返回第1步重试
/*
 set: id(1~10240)
 set: 从0～100000
 */


// json配置格式的路由表
/*
{
  "version":1,
  "sets": [
    "set_id":1,
    "set_name":"0-100",
    "range": {
      "id":0,
      "size":10
    },
    "allocs": [
      "alloc_id":,
      "addr":,
      "alloc_name":,
      "ranges":[
      ],
    ]
  ]
}
*/

//////////////////////////////////////////////////////////////////////////////
// 路由表关注的是AllocSvr服务节点
// 从客户端视角看：
//   从uid找到号段section，再从section找到AllocSvr
// 从构建路由表视角看：
//   由AllocSvr找到分配的Section，构成一个列表
// 可以设计如下数据结构
//

/*
// 大部分情况，一个AllocSvr里的号段大部分是连续的
// 为了减少网络传输量，将连续的号段使用SectionRange进行压缩
// 例如，1～10个号段，
// id_begin: 1, size: 10
struct SectionRange {
  uint32_t id_begin;  // 号段起始地址
  uint32_t size;      // 有多少个号段
};

// 路由节点
struct RouterNode {
  IpAddrInfo node_addr;               // 节点地址
  std::vector<SectionRange> section_ranges;  // 本节点管理的所有号段
};

// 客户端角度看的路由表：
struct RouterTable {
  uint32_t version;
  // std::list<RouterNode> node_list; // 整个集群所有allocsvr节点
};
 */
type RouterTable interface {
	GetVersion()
	SetVersion()
	Update() // when update, set version
	GetRoute()
	SetRoute()
}

// 仲裁服务的一个主要功能探测AllocSvr:
// 这里需要引入一个仲裁服务，探测AllocSvr的服务状态，决定每个uid段由哪台AllocSvr加载。
// 出于可靠性的考虑，仲裁模块并不直接操作AllocSvr，而是将加载配置写到StoreSvr持久化，
// 然后AllocSvr定期访问StoreSvr读取最新的加载配置，决定自己的加载状态。
//
// TODO: 引入第三方服务，比如etcd或zookeeper
type MediatetSvr interface {
	OnInit()
	Run()
	OnDestroy()
	//GetRouterTable()

	RegisterAllocSvr()
	UnRegisterAllocSvr()
	// 初始化路由表
	UpdateRouterTable()
}

// 把存储层StoreSvr
// StoreSvr为存储层，利用了多机NRW策略来保证数据持久化后不丢失
//
// NWR模型，把CAP的选择权交给了用户，让用户自己选择CAP中的哪两个。
//
// N代表N个副本（replication），
// W代表写入数据时至少要写入W份副本才认为成功，
// R表示读取数据时至少要读取R份副本。
// 对于R和W的选择，要求W+R > N。

// 通过配置文件加载
// 参数说明
// set_size: 整个系统里分配了多少个set
// set_idx:属于第几个set
// filepath: 存储路径
type StoreSvr interface {
	OnInit(settings *seqconf.Store) //db is local file.
	Run()
	OnDestroy()


	LoadRouterTable() (RouterTable, error)
	SaveRouterTable(table RouterTable) error
	GetMaxSeq(id int, size int) (int, Sequence, error)
	LoadMaxSeq() error
	SaveMaxSeq(id int, size int, maxSeq Sequence) error
}

