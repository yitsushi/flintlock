control "operating_system" do
    desc "Verifies the name of the operating system on the targeted host"
  
    describe os.name do
      it { should eq attribute("instances_ami_operating_system_name") }
    end
  end