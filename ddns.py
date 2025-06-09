import requests
import json
import time
import logging
import os
from typing import Dict, Any, Optional

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    filename='ddns.log'
)
logger = logging.getLogger(__name__)

class DNSPodDDNS:
    def __init__(self, login_token: str, domain: str, sub_domain: str = '@'):
        """
        初始化 DNSPod DDNS 客户端
        
        Args:
            login_token: DNSPod API Token，格式为 ID,Token
            domain: 主域名，如 example.com
            sub_domain: 子域名，默认为 @ 表示根域名
        """
        self.login_token = login_token
        self.domain = domain
        self.sub_domain = sub_domain
        self.domain_id = None
        self.record_id = None
        self.current_ip = None
        self.ip_cache_file = 'current_ip.txt'
        
        # DNSPod API 基础URL
        self.base_url = 'https://dnsapi.cn'
        
        # 检查 IP 缓存文件
        if os.path.exists(self.ip_cache_file):
            with open(self.ip_cache_file, 'r') as f:
                self.current_ip = f.read().strip()
                
        # 验证配置
        self._validate_config()

    def _validate_config(self):
        """验证配置有效性"""
        if not self.login_token or ',' not in self.login_token:
            raise ValueError("无效的 login_token 格式。应为 'ID,Token'")
        
        if not self.domain:
            raise ValueError("域名不能为空")
            
        logger.info(f"配置验证通过: 域名={self.domain}, 子域名={self.sub_domain}")

    def get_public_ip(self) -> Optional[str]:
        """获取当前公网 IP 地址"""
        try:
            # 使用多个服务获取公网 IP，提高可靠性
            services = [
            'https://ip.3322.net',        # 3322 IP查询服务
            'https://ipinfo.io/ip',       # IPInfo服务
            'https://icanhazip.com',      # 简单的纯文本IP服务
            'https://v4.ident.me',        # 强制IPv4服务（避免返回IPv6）
            'https://checkip.amazonaws.com', # AWS检查IP服务
            'https://ifconfig.co/ip',     # ifconfig服务
            'https://whatismyip.akamai.com',  # Akamai官方IP查询
            'https://ident.me',           # 支持IPv6的服务
            'http://ip.42.pl/raw',        # 纯文本IP服务
            'https://ipecho.net/plain'    # Echo服务
            ]
            
            for service in services:
                try:
                    headers = {'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'}
                    response = requests.get(service, headers=headers, timeout=5)
                    if response.status_code == 200:
                        return response.text.strip()
                except Exception as e:
                    logger.warning(f"获取公网 IP 失败 ({service}): {str(e)}")
            
            logger.error("所有 IP 查询服务均失败")
            return None
        except Exception as e:
            logger.error(f"获取公网 IP 时发生异常: {str(e)}")
            return None

    def get_domain_id(self) -> Optional[str]:
        """获取域名 ID"""
        if self.domain_id:
            return self.domain_id
            
        url = f"{self.base_url}/Domain.List"
        data = {
            'login_token': self.login_token,
            'format': 'json'
        }
        
        try:
            response = requests.post(url, data=data)
            result = response.json()
            
            if result.get('status', {}).get('code') == '1':
                for domain in result.get('domains', []):
                    if domain.get('name') == self.domain:
                        self.domain_id = domain.get('id')
                        logger.info(f"获取域名 ID 成功: {self.domain_id}")
                        return self.domain_id
                logger.error(f"未找到域名: {self.domain}，API返回: {result}")
            else:
                logger.error(f"获取域名 ID 失败: {result.get('status', {}).get('message')}，完整响应: {result}")
                
        except Exception as e:
            logger.error(f"获取域名 ID 时发生异常: {str(e)}")
            
        return None

    def get_record_id(self) -> Optional[str]:
        """获取记录 ID"""
        if self.record_id:
            return self.record_id
            
        domain_id = self.get_domain_id()
        if not domain_id:
            return None
            
        url = f"{self.base_url}/Record.List"
        data = {
            'login_token': self.login_token,
            'format': 'json',
            'domain_id': domain_id,
            'sub_domain': self.sub_domain
        }
        
        try:
            response = requests.post(url, data=data)
            result = response.json()
            
            if result.get('status', {}).get('code') == '1':
                for record in result.get('records', []):
                    if record.get('name') == self.sub_domain and record.get('type') == 'A':
                        self.record_id = record.get('id')
                        logger.info(f"获取记录 ID 成功: {self.record_id}")
                        return self.record_id
                logger.error(f"未找到记录: {self.sub_domain}.{self.domain}，API返回: {result}")
            else:
                logger.error(f"获取记录 ID 失败: {result.get('status', {}).get('message')}，完整响应: {result}")
                
        except Exception as e:
            logger.error(f"获取记录 ID 时发生异常: {str(e)}")
            
        return None

    def update_record(self, ip: str) -> bool:
        """更新 DNS 记录"""
        domain_id = self.get_domain_id()
        record_id = self.get_record_id()
        
        if not domain_id or not record_id:
            return False
            
        url = f"{self.base_url}/Record.Modify"
        data = {
            'login_token': self.login_token,
            'format': 'json',
            'domain_id': domain_id,
            'record_id': record_id,
            'sub_domain': self.sub_domain,
            'record_type': 'A',
            'record_line': '默认',
            'value': ip
        }
        
        try:
            response = requests.post(url, data=data)
            result = response.json()
            
            if result.get('status', {}).get('code') == '1':
                logger.info(f"成功更新记录: {self.sub_domain}.{self.domain} -> {ip}")
                self.current_ip = ip
                self._save_current_ip(ip)
                return True
            else:
                logger.error(f"更新记录失败: {result.get('status', {}).get('message')}，完整响应: {result}")
                
        except Exception as e:
            logger.error(f"更新记录时发生异常: {str(e)}")
            
        return False

    def _save_current_ip(self, ip: str):
        """保存当前 IP 到缓存文件"""
        try:
            with open(self.ip_cache_file, 'w') as f:
                f.write(ip)
        except Exception as e:
            logger.warning(f"保存 IP 缓存文件失败: {str(e)}")

    def run(self, check_interval: int = 300):
        """运行 DDNS 服务
        
        Args:
            check_interval: 检查间隔时间（秒）
        """
        logger.info(f"DNSPod DDNS 服务启动 - {self.sub_domain}.{self.domain}")
        
        while True:
            try:
                new_ip = self.get_public_ip()
                if not new_ip:
                    logger.warning("无法获取公网 IP，将在下次检查时重试")
                    time.sleep(check_interval)
                    continue
                    
                # 检查 IP 是否变化
                if new_ip != self.current_ip:
                    logger.info(f"检测到 IP 变化: {self.current_ip} -> {new_ip}")
                    if self.update_record(new_ip):
                        logger.info(f"IP 更新成功: {new_ip}")
                    else:
                        logger.error("IP 更新失败")
                else:
                    logger.debug(f"IP 未变化: {new_ip}")
                    
            except Exception as e:
                logger.error(f"运行 DDNS 检查时发生异常: {str(e)}")
                
            # 等待下一次检查
            time.sleep(check_interval)

if __name__ == "__main__":
    # 配置信息
    CONFIG = {
        'login_token': '496338,c24fef8c0d8b29b1f7f4fce43194fe39',  # DNSPod API Token
        'domain': 'stnts.eu.org',           # 主域名
        'sub_domain': 'www',               # 子域名
        'check_interval': 300              # 检查间隔（秒）
    }
    
    # 验证 API Token 格式
    if not CONFIG['login_token'] or ',' not in CONFIG['login_token']:
        print("错误: 无效的 login_token 格式。应为 'ID,Token'")
        exit(1)
    
    # 创建并运行 DDNS 客户端
    try:
        ddns_client = DNSPodDDNS(
            login_token=CONFIG['login_token'],
            domain=CONFIG['domain'],
            sub_domain=CONFIG['sub_domain']
        )
        ddns_client.run(check_interval=CONFIG['check_interval'])
    except ValueError as ve:
        logger.error(f"配置错误: {str(ve)}")
        print(f"配置错误: {str(ve)}")
    except KeyboardInterrupt:
        logger.info("DDNS 服务已停止")
        print("DDNS 服务已停止")
    except Exception as e:
        logger.error(f"启动服务时发生未知错误: {str(e)}")
        print(f"启动服务时发生未知错误: {str(e)}")